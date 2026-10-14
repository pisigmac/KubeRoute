class TrafficPredictor:
    """Exponential smoothing on request_rate_1m. No per-request model fit."""

    def predict(self, history: list[dict]) -> float:
        if not history:
            return 100.0
        if len(history) < 6:
            last = history[-1].get("request_rate_1m", history[-1].get("value", 100.0))
            return round(float(last) * 1.1, 6)

        alpha = 0.4
        level = float(history[0].get("request_rate_1m", 0))
        for h in history[1:]:
            level = alpha * float(h.get("request_rate_1m", 0)) + (1 - alpha) * level
        return float(level)
