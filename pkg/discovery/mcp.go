package discovery

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// ServerSpec is one MCP server entry from a client configuration file.
type ServerSpec struct {
	Name     string            `json:"-"`
	Type     string            `json:"type"`
	Command  string            `json:"command"`
	Args     []string          `json:"args"`
	Env      map[string]string `json:"env"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Cwd      string            `json:"cwd"`
	Disabled bool              `json:"disabled"`
}

// Transport is "stdio" for launched servers and "http" for remote ones.
func (s ServerSpec) Transport() string {
	switch {
	case s.Command != "":
		return "stdio"
	}
	return out
}

// UserConfigs returns the user-level MCP configs present on this machine.
func UserConfigs() []ConfigFile {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var desktop string
	switch runtime.GOOS {
	case "darwin":
		desktop = filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			desktop = filepath.Join(appData, "Claude", "claude_desktop_config.json")
		}
	default:
		desktop = filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}
	candidates := []ConfigFile{
		{Path: desktop, Agent: "claude-desktop", Scope: "user"},
		{Path: filepath.Join(home, ".cursor", "mcp.json"), Agent: "cursor", Scope: "user"},
		{Path: filepath.Join(home, ".claude.json"), Agent: "claude-code", Scope: "user"},
	}
	var out []ConfigFile
	for _, c := range candidates {
		if fileExists(c.Path) {
			out = append(out, c)
		}
	}
	return out
}

// comments and trailing commas that VS Code allows.
func ParseConfig(path string) ([]ServerSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var shape struct {
		MCPServers map[string]ServerSpec `json:"mcpServers"`
		Servers    map[string]ServerSpec `json:"servers"`
	}
	if err := json.Unmarshal(stripJSONC(data), &shape); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	merged := map[string]ServerSpec{}
	for name, s := range shape.Servers {
		s.Name = name
		merged[name] = s
	}
	for name, s := range shape.MCPServers {
		s.Name = name
		merged[name] = s
	}
	out := make([]ServerSpec, 0, len(merged))
	for _, s := range merged {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// stripJSONC removes // and /* */ comments and trailing commas outside strings.
func stripJSONC(in []byte) []byte {
	out := make([]byte, 0, len(in))
	inString, escaped := false, false
	for i := 0; i < len(in); i++ {
		c := in[i]
		if inString {
			out = append(out, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch {
		case c == '"':
			inString = true
			out = append(out, c)
		case c == '/' && i+1 < len(in) && in[i+1] == '/':
			for i < len(in) && in[i] != '\n' {
				i++
			}
			if i < len(in) {
				out = append(out, '\n')
			}
		case c == '/' && i+1 < len(in) && in[i+1] == '*':
			i += 2
			for i+1 < len(in) && !(in[i] == '*' && in[i+1] == '/') {
				i++
			}
			i++ // land on the closing '/'; the loop increment skips it
		case c == ',':
			j := i + 1
			for j < len(in) && (in[j] == ' ' || in[j] == '\t' || in[j] == '\n' || in[j] == '\r') {
				j++
			}
			if j < len(in) && (in[j] == '}' || in[j] == ']') {
				continue // trailing comma
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
	}
	return out
}

// ---- Extraction: Agent → MCP Server → Resource ---------------------------

func (d *discoverer) addConfig(cf ConfigFile) error {
	specs, err := ParseConfig(cf.Path)
	if err != nil {
		if cf.Scope == "user" {
			d.logf("skipping %s: %v", cf.Path, err)
			return nil
		}
		return err
	}
	display := d.displayPath(cf.Path)
	agentID := "agent:" + cf.Agent
	agent := d.g.AddNode(&graph.Node{ID: agentID, Name: AgentName(cf.Agent), Type: graph.NodeAgent})
	appendMeta(agent, "configs", display)

	for _, s := range specs {
		if s.Disabled {
			continue
		}
		serverID := d.serverID(s, cf)
		meta := map[string]string{graph.MetaSource: display, graph.MetaTransport: s.Transport()}
		if s.Command != "" {
			meta["command"] = s.Command
		}
		if s.URL != "" {
			meta[graph.MetaURL] = RedactURL(s.URL)
		}
		d.g.AddNode(&graph.Node{ID: serverID, Name: s.Name, Type: graph.NodeMCPServer, Metadata: meta})
		d.g.AddEdge(graph.Edge{SourceID: agentID, TargetID: serverID, Relation: "connects"})
		d.inferFromSpec(serverID, s)

		if cf.Scope != "user" {
			d.recordSourceDirs(serverID, s, filepath.Dir(cf.Path))
		}
		if _, seen := d.specs[serverID]; !seen && s.Command != "" {
			spec := s
			spec.Cwd = resolveDir(filepath.Dir(cf.Path), s.Cwd, cf.Scope != "user")
			d.specs[serverID] = spec
		}
	}
	return nil
}

// serverID keys servers by name, so the same server configured for two
// clients is one node. The same name with a different command gets its own.
func (d *discoverer) serverID(s ServerSpec, cf ConfigFile) string {
	id := "server:" + s.Name
	existing, ok := d.g.Nodes[id]
	if !ok {
		return id
	}
	server := d.g.Nodes[serverID]
	var secrets []string

	for _, name := range sortedKeys(s.Env) {
		if res, ok := CredentialResource(name); ok {
			d.addResource(serverID, "resource:"+res, res, "credential", graph.PermWrite, true, map[string]string{"via": name})
		}
		if LooksLikeHardcodedSecret(name, s.Env[name]) {
			secrets = append(secrets, "env "+name)
		}
	}
	for _, name := range sortedKeys(s.Headers) {
		if LooksLikeHardcodedSecret(name, s.Headers[name]) {
			secrets = append(secrets, "header "+name)
		}
	}
	if s.URL != "" && URLHasCredentials(s.URL) {
		secrets = append(secrets, "url")
	}

	for _, arg := range s.Args {
		if isDSN, hasPassword := DSNArg(arg); isDSN {
			d.addResource(serverID, "resource:database", "database", "connection-string", graph.PermWrite, true, nil)
			if hasPassword {
				secrets = append(secrets, "connection-string argument")
			}
		}
	}
	for _, p := range PathArgs(s.Args) {
		meta := map[string]string{"path": p}
		if IsBroadPath(p) {
			meta[graph.MetaBroadPath] = "true"
		}
		d.addResource(serverID, "resource:path:"+p, p, "filesystem", graph.PermWrite, true, meta)
	}

	if pkg, unpinned := UnpinnedPackage(s.Command, s.Args); unpinned {
		server.Metadata[graph.MetaUnpinnedPackage] = pkg
	}
	if len(secrets) > 0 {
		server.Metadata[graph.MetaHardcodedSecrets] = strings.Join(secrets, ", ")
	}
}

// recordSourceDirs remembers which repository directories hold a server's
// code (e.g. `python server.py`), so tools found there attach to that server.
func (d *discoverer) recordSourceDirs(serverID string, s ServerSpec, configDir string) {
	base := resolveDir(configDir, s.Cwd, true)
	for _, arg := range s.Args {
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		p := arg
		if !filepath.IsAbs(p) {
		if err != nil {
			continue
		}
		isScript := !info.IsDir() && scriptExts[strings.ToLower(filepath.Ext(p))]
		isCodeDir := info.IsDir() && isInterpreter(s.Command)
		if !isScript && !isCodeDir {
			continue
		}
		rel, err := filepath.Rel(d.root, p)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if !info.IsDir() {
			rel = filepath.Dir(rel)
		}
		d.sourceDirs[serverID] = append(d.sourceDirs[serverID], filepath.ToSlash(rel))
	}
}

// ---- Heuristics ------------------------------------------------------------

// ToolAnnotations are the behaviour hints an MCP server reports per tool.
// They are self-reported, so Raaya treats them as hints rather than proof.
type ToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    *bool  `json:"readOnlyHint,omitempty"`
	DestructiveHint *bool  `json:"destructiveHint,omitempty"`
	IdempotentHint  *bool  `json:"idempotentHint,omitempty"`
	OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`
}

// Values for graph.MetaPermissionSource.
const (
	permFromReadOnlyHint    = "readOnlyHint"
	permFromDestructiveHint = "destructiveHint"
	permFromName            = "name-heuristic"
	permFromOverride        = "raaya-override"
)

var execName = regexp.MustCompile(`(?i)(^|[_\-.\s])(shell|exec|execute|bash|terminal|command|cmd|eval|subprocess|run_script|powershell)($|[_\-.\s])`)

// InferToolPermission maps a tool's annotations and name to a permission and
// returns how it was decided. Per the MCP spec, a tool without annotations is
// presumed to change state, so the fallback is WRITE.
func InferToolPermission(name string, a *ToolAnnotations) (graph.PermissionLevel, string) {
	if a != nil && a.ReadOnlyHint != nil && *a.ReadOnlyHint {
		return graph.PermRead, permFromReadOnlyHint
	}
	if execName.MatchString(name) {
		return graph.PermExecute, permFromName
	}
	if a == nil || (a.ReadOnlyHint == nil && a.DestructiveHint == nil) {
		return graph.PermWrite, graph.PermissionAssumed
	}
	return graph.PermWrite, permFromDestructiveHint
}

type credentialRule struct {
	pattern  *regexp.Regexp
	resource string
}

var credentialRules = []credentialRule{
	{regexp.MustCompile(`^(GITHUB|GH)_`), "github"},
	{regexp.MustCompile(`^GITLAB_`), "gitlab"},
	{regexp.MustCompile(`^(AWS|AMAZON)_`), "aws"},
	{regexp.MustCompile(`^(GOOGLE|GCP|GCLOUD)_`), "gcp"},
	placeholderValue = regexp.MustCompile(`(?i)^\s*$|\$\{|\{\{|^<.*>$|^(your|my|example|changeme|xxx|placeholder|todo|redacted|dummy)`)
	knownTokenPrefix = regexp.MustCompile(`^(ghp_|gho_|ghs_|ghu_|github_pat_|glpat-|sk-|sk_live_|rk_live_|xox[abpr]-|AKIA[0-9A-Z]{12}|AIza)`)
)

// LooksLikeHardcodedSecret reports whether value is a literal credential
// rather than a reference such as ${VAR}, ${env:VAR} or ${input:token}.
func LooksLikeHardcodedSecret(name, value string) bool {
	value = strings.TrimSpace(value)
	for _, scheme := range []string{"Bearer ", "bearer ", "Basic ", "basic ", "token "} {
		value = strings.TrimPrefix(value, scheme)
	}
	if placeholderValue.MatchString(value) {
		return false
	}
	if knownTokenPrefix.MatchString(value) {
		return true
	}
	upper := strings.ToUpper(name)
	if !secretName.MatchString(upper) && upper != "AUTHORIZATION" {
		return false
	}
	return len(value) >= 12
}

// URLHasCredentials reports whether a server URL embeds a password or a
// secret-looking query parameter.
func URLHasCredentials(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.User != nil {
		if pw, ok := u.User.Password(); ok && !placeholderValue.MatchString(pw) {
			return true
		}
}

// RedactURL strips credentials, query and fragment so a URL is safe to report.
func RedactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.User, u.RawQuery, u.Fragment = nil, "", ""
	return u.String()
}

var dsnPattern = regexp.MustCompile(`(?i)^(postgres(ql)?|mysql|mariadb|mongodb(\+srv)?|rediss?|sqlserver|clickhouse)://`)

// DSNArg reports whether arg is a database connection string, and whether it
// embeds a literal password.
func DSNArg(arg string) (isDSN, hasPassword bool) {
	if !dsnPattern.MatchString(arg) {
		return false, false
	}
	u, err := url.Parse(arg)
	if err != nil || u.User == nil {
		return true, false
	}
	pw, ok := u.User.Password()
	return true, ok && pw != "" && !placeholderValue.MatchString(pw)
}

// UnpinnedPackage returns the package a server is launched from via npx,
// bunx, pnpx or uvx, and whether that package lacks a pinned version.
func UnpinnedPackage(command string, args []string) (string, bool) {
	runner := strings.ToLower(filepath.Base(command))
	runner = strings.TrimSuffix(strings.TrimSuffix(runner, ".exe"), ".cmd")
	switch runner {
	case "npx", "bunx", "pnpx":
		pkg := packageArg(args, []string{"-p", "--package"}, []string{"--registry", "--cache", "-c", "--call"})
			return pkg, true
		}
		v := name[at+1:]
		return pkg, v == "" || v == "latest" || v == "next"
	case "uvx":
		pkg := packageArg(args, []string{"--from"}, []string{"--python", "-p", "--with", "--index", "--index-url"})
		if pkg == "" {
			return "", false
		}
		return pkg, !strings.Contains(pkg, "==") && !strings.Contains(pkg, "@")
	}
	return "", false
}

// packageArg returns the package named in args: the value of a pkgFlag, or
// the first positional argument (skipping valueFlags and their values).
func packageArg(args, pkgFlags, valueFlags []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		for _, f := range pkgFlags {
			if a == f && i+1 < len(args) {
				return args[i+1]
			}
			if strings.HasPrefix(a, f+"=") {
				return strings.TrimPrefix(a, f+"=")
			}
		}
		skip := false
		for _, f := range valueFlags {
			if a == f {
				i++
				skip = true
				break
			}
		}
		if skip || strings.HasPrefix(a, "-") {
			continue
		}
		return a
	}
	return ""
}

var (
	windowsDrive = regexp.MustCompile(`^[A-Za-z]:([\\/]|$)`)
	scriptExts   = map[string]bool{".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".py": true, ".jar": true, ".rb": true, ".sh": true}
	interpreters = map[string]bool{
		"python": true, "python3": true, "node": true, "deno": true, "bun": true,
		"uv": true, "go": true, "tsx": true, "ts-node": true, "ruby": true,
	}
)

// PathArgs returns arguments that look like filesystem paths handed to a
// server (for example the allowed directories of a filesystem server).
		}
		switch {
		case a == "." || a == "~",
			strings.HasPrefix(a, "/"), strings.HasPrefix(a, "~/"),
			strings.HasPrefix(a, "./"), strings.HasPrefix(a, "../"),
			strings.HasPrefix(a, "$HOME"), strings.HasPrefix(a, "${HOME}"),
			windowsDrive.MatchString(a):
			out = append(out, a)
		}
	}
	return out
}

// IsBroadPath reports whether p is the filesystem root or a home directory.
func IsBroadPath(p string) bool {
	clean := strings.TrimRight(p, `/\`)
	switch clean {
	case "", "~", "$HOME", "${HOME}", "%USERPROFILE%":
		return true
	}
	if len(clean) == 2 && clean[1] == ':' {
		return true // C:\
	}
	if home, err := os.UserHomeDir(); err == nil && filepath.Clean(p) == filepath.Clean(home) {
		return true
	}
	return false
}

func isInterpreter(command string) bool {
	base := strings.ToLower(filepath.Base(command))
	return interpreters[strings.TrimSuffix(base, ".exe")]
}
