from ai_engine.models.cost import CostOptimizer
from ai_engine.models.traffic import TrafficPredictor

CONSTRAINTS = ["skip kube-system", "skip DaemonSet", "spot_share <= 60%", "fit CPU and memory"]


def _proposed_patch(kind: str, ns: str, name: str, node: str, op: str) -> dict:
    patch = {
        "op": op,
        "kind": kind,
        "namespace": ns,
        "name": name,
    }
    if op == "unpin":
        patch["nodeAffinity"] = None
        return patch
    patch["nodeAffinity"] = {
        "requiredDuringSchedulingIgnoredDuringExecution": {
            "nodeSelectorTerms": [
                {
                    "matchExpressions": [
                        {
                            "key": "kubernetes.io/hostname",
                            "operator": "In",
                            "values": [node],
                        }
                    ]
                }
            ]
        }
    }
    return patch


class DecisionPipeline:
    async def process(self, metrics: dict, db_pool=None) -> list[dict]:
        decisions = []
        cluster_id = metrics.get("cluster_id")
        nodes = metrics.get("nodes") or []
        workloads = metrics.get("workloads") or []

        if nodes and workloads:
            opt = CostOptimizer()
            assignments = opt.optimize(nodes, workloads)
            by_name = {n["name"]: n for n in nodes}
            wl_by_name = {w["name"]: w for w in workloads}
            for a in assignments:
                w = wl_by_name.get(a["workload"])
                if not w:
                    continue
                current = w.get("current_node") or w.get("node") or ""
                if current == a["node"]:
                    continue
                old_cost = by_name.get(current, {}).get("cost_per_hour", 0) if current else 0
                new_cost = a.get("cost_per_hour", 0)
                savings_hr = max(0.0, float(old_cost) - float(new_cost))
                conf = 0.55
                if old_cost:
                    conf = min(0.99, 0.5 + (savings_hr / float(old_cost)) * 0.5)
                ns = w.get("namespace") or "default"
                kind = w.get("kind") or "Deployment"
                candidates = opt.explain_candidates(nodes, w, a["node"])
                decisions.append({
                    "cluster_id": str(cluster_id) if cluster_id is not None else "",
                    "decision_type": "placement",
                    "reason": f"Move {ns}/{a['workload']} from {current or 'unscheduled'} to {a['node']}",
                    "confidence": round(conf, 3),
                    "params": {
                        "target_node": a["node"],
                        "workload": a["workload"],
                        "namespace": ns,
                        "kind": kind,
                        "current_node": current,
                        "op": "pin",
                    },
                    "predicted_cost_savings": round(savings_hr * 730, 2),
                    "solver_trace": {
                        "objective": "minimize_hourly_cost",
                        "chosen": a["node"],
                        "current": current,
                        "candidates": candidates,
                        "constraints": CONSTRAINTS,
                        "checks": [
                            "fits CPU/mem on target",
                            "not kube-system",
                            "not DaemonSet",
                        ],
                        "blast_radius": {
                            "kind": kind,
                            "namespace": ns,
                            "name": a["workload"],
                            "workloads": 1,
                        },
                        "savings_per_hour": round(savings_hr, 6),
                    },
                    "proposed_patch": _proposed_patch(kind, ns, a["workload"], a["node"], "pin"),
                })

        if "traffic" in metrics:
            tp = TrafficPredictor()
            pred = tp.predict(metrics["traffic"])
            current = metrics.get("current_rps", 0)
            if current and pred > current * 1.5:
                decisions.append({
                    "cluster_id": cluster_id,
                    "decision_type": "scaling",
                    "reason": f"Predicted RPS {pred:.0f} exceeds threshold",
                    "confidence": min(0.99, round(pred / max(current, 1) - 0.5, 3)),
                    "params": {"replicas": "+2"},
                    "predicted_cost_savings": 0,
                })

        return decisions
