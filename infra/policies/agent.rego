package kuberoute.agent

# Deny evicting system pods
deny[msg] {
    input.decision.decision_type == "placement"
    input.workload.namespace == "kube-system"
    msg := "Cannot move workloads in kube-system"
}

# Deny draining more than 30% of cluster capacity
deny[msg] {
    input.decision.decision_type == "failover"
    input.drain_percentage > 30
    msg := "Cannot drain more than 30% of cluster capacity"
}

# Deny cross-namespace moves without label
deny[msg] {
    input.decision.decision_type == "placement"
    input.workload.namespace != input.target.namespace
    not input.workload.labels["kuberoute.io/allow-cross-namespace"]
    msg := "Cross-namespace moves require kuberoute.io/allow-cross-namespace label"
}
