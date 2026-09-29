package reporter

import (
	"bytes"
	"fmt"
	"text/template"

	"raaya/pkg/analysis"
)

const prCommentTemplate = `## 🛡️ Raaya Security Surface & Capability Diff

{{if .HasRegressions}}
> ⚠️ **Warning:** Security surface expansion or permission escalation detected in this PR.
{{else}}
> ✅ No dangerous capability expansions detected.
{{end}}

### 📊 Summary
* **Added Entities:** {{len .AddedNodes}}
* **New Capability Edges:** {{len .AddedEdges}}
* **Permission Escalations:** {{len .Escalations}}

{{if .Escalations}}
### 🚨 Permission Escalations
| Source | Target | New Permission |
| :--- | :--- | :--- |
{{range .Escalations}}| ` + "`{{.SourceID}}`" + ` | ` + "`{{.TargetID}}`" + ` | **{{.Permission}}** |
{{end}}
{{end}}

{{if .AddedEdges}}
### 🔗 Newly Introduced Capabilities
| Source | Capability | Target |
| :--- | :--- | :--- |
{{range .AddedEdges}}| ` + "`{{.SourceID}}`" + ` | {{.Capability}} | ` + "`{{.TargetID}}`" + ` |
{{end}}
{{end}}
`

// GeneratePRComment builds formatted Markdown content ready for GitHub API post-back.
func GeneratePRComment(diff *analysis.CapabilityDiff) (string, error) {
	tmpl, err := template.New("pr_comment").Parse(prCommentTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, diff); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
