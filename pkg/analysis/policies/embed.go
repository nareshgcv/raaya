// Package policies bundles Raaya's default Rego policy.
package policies

import _ "embed"

// Default is the bundled policy, evaluated in builds with -tags rego.
//
//go:embed default.rego
var Default string
