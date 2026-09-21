package ape.security

default allow = false

# Identify high-risk tool connections
violations[msg] {
    some i
    edge := input.edges[i]
    edge.type == "ACCESSES"
    
    target_node := input.nodes[edge.target_id]
    target_node.type == "RESOURCE"
    target_node.metadata.sensitivity == "HIGH"
    
    msg := sprintf("High-risk resource connection detected: %s -> %s", [edge.source_id, edge.target_id])
}

allow {
    count(violations) == 0
}
