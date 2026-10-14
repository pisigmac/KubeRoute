from ai_engine.models.placement import PlacementModel


def test_placement_basic():
    pm = PlacementModel()
    workload = {"name": "app", "cpu_request": 1, "memory_request": 2_000_000_000}
    nodes = [
        {"name": "n1", "capacity_cpu": 4, "capacity_memory": 8_000_000_000, "allocated_cpu": 3, "allocated_memory": 6_000_000_000, "cost_per_hour": 0.05},
        {"name": "n2", "capacity_cpu": 4, "capacity_memory": 8_000_000_000, "allocated_cpu": 1, "allocated_memory": 2_000_000_000, "cost_per_hour": 0.03},
    ]
    result = pm.place(workload, nodes)
    assert result["node"] == "n2"
    assert 0.5 <= result["confidence"] <= 0.99


def test_placement_no_capacity():
    pm = PlacementModel()
    workload = {"name": "app", "cpu_request": 8, "memory_request": 32_000_000_000}
    nodes = [
        {"name": "n1", "capacity_cpu": 4, "capacity_memory": 16_000_000_000, "allocated_cpu": 0, "allocated_memory": 0, "cost_per_hour": 0.05},
    ]
    result = pm.place(workload, nodes)
    assert result["node"] is None
