package raaya

# Default Raaya policy, evaluated when built with -tags rego.
# Add organisation policies with --policy; every file must use `package raaya`
# and add to `deny`. Rego v1 syntax (`contains`, `if`) is required.
#
# input.reach[agent][node] is the effective permission an agent has on a node.

deny contains finding if {
	some agent, node
	input.reach[agent][node] == "ADMIN"
	input.graph.nodes[node].type == "RESOURCE"
	finding := {
		"rule_id": "POLICY-ADMIN-RESOURCE",
		"severity": "HIGH",
		"agent_id": agent,
		"node_id": node,
		"msg": sprintf("%s has ADMIN on resource %s", [agent, node]),
	}
}
