import math
from typing import Optional


class PlacementModel:
    def place(self, workload: dict, nodes: list[dict]) -> dict:
        ranked = []
        w_cpu = workload.get("cpu_request", 0.5)
        w_mem = workload.get("memory_request", 512 * 1024 * 1024)

        for node in nodes:
            cpu_avail = node.get("capacity_cpu", 4) - node.get("allocated_cpu", 0)
            mem_avail = node.get("capacity_memory", 16 * 1024 * 1024 * 1024) - node.get("allocated_memory", 0)
            if cpu_avail < w_cpu or mem_avail < w_mem:
                continue
            cost = node.get("cost_per_hour", 0.05)
            headroom = (cpu_avail / max(node.get("capacity_cpu", 1), 0.001)) * 0.5 + (
                mem_avail / max(node.get("capacity_memory", 1), 0.001)
            ) * 0.5
            score = headroom - cost * 2
            ranked.append((score, node))

        if not ranked:
            return {"node": None, "confidence": 0.0, "reason": "No node has sufficient capacity"}

        ranked.sort(key=lambda x: x[0], reverse=True)
        best_score, best_node = ranked[0]
        second = ranked[1][0] if len(ranked) > 1 else None
        confidence = _confidence(best_score, second, len(ranked))
        return {
            "node": best_node["name"],
            "confidence": confidence,
            "reason": f"Selected {best_node['name']} based on cost and headroom",
        }


def _confidence(best: float, second: Optional[float], n: int) -> float:
    if n <= 0:
        return 0.0
    if n == 1 or second is None:
        return 0.55
    margin = max(best - second, 0.0)
    return round(min(0.99, 0.5 + 0.49 * math.tanh(margin * 3)), 3)
