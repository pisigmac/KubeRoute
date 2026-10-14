from ai_engine.models.cost import CostOptimizer


def test_cost_optimizer_basic():
    opt = CostOptimizer()
    nodes = [
        {"name": "n1", "capacity_cpu": 4, "capacity_memory": 16_000_000_000, "cost_per_hour": 0.05, "spot_instance": False},
        {"name": "n2", "capacity_cpu": 4, "capacity_memory": 16_000_000_000, "cost_per_hour": 0.02, "spot_instance": True},
    ]
    workloads = [
        {"name": "w1", "cpu_request": 1, "memory_request": 2_000_000_000},
    ]
    result = opt.optimize(nodes, workloads)
    assert len(result) == 1
    assert result[0]["workload"] == "w1"


def test_explain_candidates_marks_current_and_chosen():
    opt = CostOptimizer()
    nodes = [
        {
            "name": "n1",
            "capacity_cpu": 4,
            "capacity_memory": 16_000_000_000,
            "allocated_cpu": 1,
            "allocated_memory": 1_000_000_000,
            "cost_per_hour": 0.05,
        },
        {
            "name": "n2",
            "capacity_cpu": 4,
            "capacity_memory": 16_000_000_000,
            "allocated_cpu": 0,
            "allocated_memory": 0,
            "cost_per_hour": 0.02,
        },
    ]
    workload = {
        "name": "w1",
        "cpu_request": 0.5,
        "memory_request": 512_000_000,
        "current_node": "n1",
    }
    rows = opt.explain_candidates(nodes, workload, "n2")
    by_node = {r["node"]: r for r in rows}
    assert by_node["n2"]["verdict"] == "chosen"
    assert by_node["n1"]["verdict"] == "current"


def test_cost_optimizer_spot_constraint():
    opt = CostOptimizer()
    nodes = [
        {"name": "n1", "capacity_cpu": 4, "capacity_memory": 16_000_000_000, "cost_per_hour": 0.01, "spot_instance": True},
        {"name": "n2", "capacity_cpu": 4, "capacity_memory": 16_000_000_000, "cost_per_hour": 0.05, "spot_instance": False},
    ]
    workloads = [
        {"name": "w1", "cpu_request": 0.5, "memory_request": 1_000_000_000},
        {"name": "w2", "cpu_request": 0.5, "memory_request": 1_000_000_000},
        {"name": "w3", "cpu_request": 0.5, "memory_request": 1_000_000_000},
        {"name": "w4", "cpu_request": 0.5, "memory_request": 1_000_000_000},
        {"name": "w5", "cpu_request": 0.5, "memory_request": 1_000_000_000},
    ]
    result = opt.optimize(nodes, workloads)
    spot_count = sum(1 for r in result if r["node"] == "n1")
    assert spot_count <= 3  # 60% of 5
