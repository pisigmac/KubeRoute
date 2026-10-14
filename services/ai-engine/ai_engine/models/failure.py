class FailurePredictor:
    """Rule-based at-risk detector. Not a trained model — labels are the rules."""

    def __init__(self):
        self.threshold = 0.7

    def predict(self, node_metrics: list[dict]) -> list[dict]:
        at_risk = []
        for nm in node_metrics:
            disk_errors = nm.get("disk_errors", 0)
            disk_util = nm.get("disk_util", 0)
            if disk_errors <= 5 and disk_util <= 90:
                continue
            if disk_errors > 5 and disk_util > 90:
                prob = 0.95
            elif disk_errors > 5:
                prob = 0.85
            else:
                prob = 0.75
            at_risk.append({
                "node": nm.get("node_name"),
                "failure_probability": prob,
                "rule": "disk_errors>5 or disk_util>90",
            })
        return at_risk
