import asyncio
import json
import logging
import os
from contextlib import asynccontextmanager

import asyncpg
from fastapi import FastAPI
from nats.aio.client import Client as NATS
from redis.asyncio import Redis

from ai_engine.models.cost import CostOptimizer
from ai_engine.models.failure import FailurePredictor
from ai_engine.models.placement import PlacementModel
from ai_engine.models.traffic import TrafficPredictor
from ai_engine.pipelines.decision import DecisionPipeline

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("ai-engine")

redis: Redis | None = None
nats_client: NATS | None = None
db_pool: asyncpg.Pool | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global redis, nats_client, db_pool

    db_pool = await asyncpg.create_pool(
        os.getenv("DATABASE_URL", "postgres://kuberoute:kuberoute_dev@localhost:5432/kuberoute")
    )
    redis = Redis.from_url(f"redis://{os.getenv('REDIS_URL', 'localhost:6379')}")
    nats_client = NATS()
    await nats_client.connect(os.getenv("NATS_URL", "nats://localhost:4222"))

    asyncio.create_task(decision_consumer())

    yield

    await db_pool.close()
    await redis.close()
    await nats_client.close()


app = FastAPI(title="Kuberoute AI Engine", version="1.0.0", lifespan=lifespan)

pipeline = DecisionPipeline()


@app.get("/health")
async def health():
    return {"status": "ok", "models_loaded": True}


@app.post("/predict/traffic")
async def predict_traffic(cluster_id: str, history: list[dict]):
    model = TrafficPredictor()
    result = model.predict(history)
    return {"cluster_id": cluster_id, "predicted_rps_30m": result}


@app.post("/predict/cost")
async def optimize_cost(cluster_id: str, nodes: list[dict], workloads: list[dict]):
    optimizer = CostOptimizer()
    result = optimizer.optimize(nodes, workloads)
    return {"cluster_id": cluster_id, "assignments": result}


@app.post("/predict/failure")
async def predict_failure(cluster_id: str, node_metrics: list[dict]):
    predictor = FailurePredictor()
    result = predictor.predict(node_metrics)
    return {"cluster_id": cluster_id, "at_risk_nodes": result}


@app.post("/predict/placement")
async def predict_placement(cluster_id: str, workload: dict, nodes: list[dict]):
    model = PlacementModel()
    result = model.place(workload, nodes)
    return {"cluster_id": cluster_id, "placement": result}


@app.post("/optimize/snapshot")
async def optimize_snapshot(payload: dict):
    decisions = await pipeline.process(payload, db_pool)
    return {"decisions": decisions}


async def decision_consumer():
    if nats_client is None:
        return
    sub = await nats_client.subscribe("metrics.raw")
    async for msg in sub.messages:
        try:
            data = json.loads(msg.data)
            decisions = await pipeline.process(data, db_pool)
            for d in decisions:
                await nats_client.publish("decisions.new", json.dumps(d).encode())
        except Exception as e:
            logger.error(f"Decision pipeline error: {e}")
