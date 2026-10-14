import math

import pulp


class CostOptimizer:
    def optimize(self, nodes: list[dict], workloads: list[dict]) -> list[dict]:
        prob = pulp.LpProblem("cost_optimization", pulp.LpMinimize)

        assignments = {}
        for w in workloads:
            for n in nodes:
                var_name = f"{w['name']}_{n['name']}"
                assignments[(w["name"], n["name"])] = pulp.LpVariable(var_name, cat="Binary")

        prob += pulp.lpSum(
            assignments[(w["name"], n["name"])] * n.get("cost_per_hour", 0.05)
            for w in workloads for n in nodes
        )

        for w in workloads:
            prob += pulp.lpSum(assignments[(w["name"], n["name"])] for n in nodes) == 1

        for n in nodes:
            cpu_cap = n.get("capacity_cpu", 4)
            mem_cap = n.get("capacity_memory", 16 * 1024 * 1024 * 1024)
            prob += pulp.lpSum(
                assignments[(w["name"], n["name"])] * w.get("cpu_request", 0.5)
                for w in workloads
            ) <= cpu_cap
            prob += pulp.lpSum(
                assignments[(w["name"], n["name"])] * w.get("memory_request", 512 * 1024 * 1024)
                for w in workloads
            ) <= mem_cap

        spot_nodes = [n for n in nodes if n.get("spot_instance")]
        if spot_nodes:
            total = len(workloads)
            prob += pulp.lpSum(
                assignments[(w["name"], n["name"])]
                for w in workloads for n in spot_nodes
            ) <= int(math.ceil(total * 0.6))

        prob.solve(pulp.PULP_CBC_CMD(msg=False))

        result = []
        for w in workloads:
            for n in nodes:
                if pulp.value(assignments[(w["name"], n["name"])]) == 1:
                    result.append({
                        "workload": w["name"],
                        "node": n["name"],
                        "cost_per_hour": n.get("cost_per_hour", 0),
                    })
        return result

    def explain_candidates(self, nodes: list[dict], workload: dict, chosen: str) -> list[dict]:
        """Per-node veto list for one assignment. Isolated fit, not a second ILP."""
        w_cpu = float(workload.get("cpu_request", 0.5))
        w_mem = float(workload.get("memory_request", 512 * 1024 * 1024))
        current = workload.get("current_node") or workload.get("node") or ""
        out = []
        for n in nodes:
            cpu_alloc = float(n.get("allocated_cpu", 0))
            mem_alloc = float(n.get("allocated_memory", 0))
            if current == n["name"]:
                cpu_alloc = max(0.0, cpu_alloc - w_cpu)
                mem_alloc = max(0.0, mem_alloc - w_mem)
            cpu_avail = float(n.get("capacity_cpu", 4)) - cpu_alloc
            mem_avail = float(n.get("capacity_memory", 16 * 1024 * 1024 * 1024)) - mem_alloc
            cost = float(n.get("cost_per_hour", 0.05))
            row = {"node": n["name"], "cost_per_hour": cost}
            if n["name"] == chosen:
                row["verdict"] = "chosen"
                row["why"] = "feasible, lowest cost under constraints"
            elif cpu_avail + 1e-9 < w_cpu:
                row["verdict"] = "rejected"
                row["why"] = f"CPU {cpu_avail:.2f} available < request {w_cpu:.2f}"
            elif mem_avail + 1e-9 < w_mem:
                row["verdict"] = "rejected"
                row["why"] = f"memory {mem_avail:.0f} available < request {w_mem:.0f}"
            elif n["name"] == current:
                row["verdict"] = "current"
                row["why"] = f"already here; more expensive than {chosen}"
            else:
                row["verdict"] = "rejected"
                row["why"] = "feasible but not optimal (cost or spot-share constraint)"
            out.append(row)
        return out
