import pytest
from ai_engine.models.traffic import TrafficPredictor


def test_traffic_predictor_basic():
    tp = TrafficPredictor()
    history = [
        {"request_rate_1m": 100, "request_rate_5m": 95, "latency_p99": 50, "error_rate": 0.01, "hour": 12, "day_of_week": 1},
        {"request_rate_1m": 110, "request_rate_5m": 100, "latency_p99": 55, "error_rate": 0.01, "hour": 12, "day_of_week": 1},
    ]
    result = tp.predict(history)
    assert isinstance(result, float)
    assert result > 0


def test_traffic_predictor_empty():
    tp = TrafficPredictor()
    result = tp.predict([])
    assert result == 100.0


def test_traffic_predictor_short_history():
    tp = TrafficPredictor()
    history = [
        {"request_rate_1m": 100, "request_rate_5m": 95, "latency_p99": 50, "error_rate": 0.01, "hour": 12, "day_of_week": 1},
    ]
    result = tp.predict(history)
    assert result == 110.0  # 100 * 1.1
