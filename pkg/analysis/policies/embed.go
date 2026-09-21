package policies

import "embed"

//go:embed *.rego
var DefaultPolicies embed.FS
