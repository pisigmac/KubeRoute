from ai_engine.models.failure import FailurePredictor


def test_failure_predictor_basic():
    fp = FailurePredictor()
    metrics = [
        {"node_name": "n1", "cpu_temp": 45, "disk_errors": 0, "memory_ecc": 0, "load_avg": 1.0, "disk_util": 40},
        {"node_name": "n2", "cpu_temp": 80, "disk_errors": 10, "memory_ecc": 2, "load_avg": 3.5, "disk_util": 95},
    ]
    result = fp.predict(metrics)
    assert len(result) >= 1
    assert result[0]["node"] == "n2"
    assert result[0]["failure_probability"] > 0.7


def test_failure_predictor_empty():
    fp = FailurePredictor()
    result = fp.predict([])
    assert result == []


def test_failure_predictor_safe_nodes():
    fp = FailurePredictor()
    metrics = [
        {"node_name": "n1", "cpu_temp": 40, "disk_errors": 0, "memory_ecc": 0, "load_avg": 0.5, "disk_util": 20},
        {"node_name": "n2", "cpu_temp": 42, "disk_errors": 0, "memory_ecc": 0, "load_avg": 0.6, "disk_util": 25},
    ]
    result = fp.predict(metrics)
    assert len(result) == 0
