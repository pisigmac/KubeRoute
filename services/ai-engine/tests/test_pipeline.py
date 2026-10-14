import asyncio

from ai_engine.pipelines.decision import DecisionPipeline


def test_pipeline_placement_from_snapshot():
    pipeline = DecisionPipeline()
    metrics = {
        "cluster_id": "c1",
        "nodes": [
            {
                "name": "n1",
                "capacity_cpu": 4,
                "capacity_memory": 16_000_000_000,
                "allocated_cpu": 1,
                "allocated_memory": 1_000_000_000,
                "cost_per_hour": 0.05,
                "spot_instance": False,
            },
            {
                "name": "n2",
                "capacity_cpu": 4,
                "capacity_memory": 16_000_000_000,
                "allocated_cpu": 0,
                "allocated_memory": 0,
                "cost_per_hour": 0.02,
                "spot_instance": True,
            },
        ],
        "workloads": [
            {
                "name": "web",
                "namespace": "default",
                "kind": "Deployment",
                "cpu_request": 0.5,
                "memory_request": 512_000_000,
                "current_node": "n1",
            },
        ],
    }
    decisions = asyncio.run(pipeline.process(metrics, None))
    assert len(decisions) == 1
    d = decisions[0]
    assert d["decision_type"] == "placement"
    assert d["params"]["target_node"] == "n2"
    assert d["params"]["current_node"] == "n1"
    assert d["params"]["op"] == "pin"
    trace = d["solver_trace"]
    assert trace["chosen"] == "n2"
    assert trace["current"] == "n1"
    verdicts = {c["node"]: c["verdict"] for c in trace["candidates"]}
    assert verdicts["n2"] == "chosen"
    assert verdicts["n1"] == "current"
    assert d["proposed_patch"]["op"] == "pin"
    assert "n2" in d["proposed_patch"]["nodeAffinity"]["requiredDuringSchedulingIgnoredDuringExecution"]["nodeSelectorTerms"][0]["matchExpressions"][0]["values"]
