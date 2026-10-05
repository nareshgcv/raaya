package discovery

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ClientVersion is reported to MCP servers during initialize.
var ClientVersion = "dev"

const (
	mcpProtocolVersion = "2025-06-18"
	maxListPages       = 50
)

// LiveTool is a tool as reported by tools/list.
type LiveTool struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Annotations *ToolAnnotations `json:"annotations"`
}

// LiveResource is a resource as reported by resources/list.
type LiveResource struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// LiveResult is what a running server reported about itself.
type LiveResult struct {
	ServerInfo string
	Tools      []LiveTool
	Resources  []LiveResource
}

// ListServer launches a stdio MCP server, performs the initialize handshake
// and lists its tools and resources.
//
// This executes the configured command. Only call it for configs you trust:
// never on pull requests from forks.
func ListServer(ctx context.Context, spec ServerSpec, timeout time.Duration) (*LiveResult, error) {
	if spec.Command == "" {
		return nil, errors.New("only stdio servers can be queried live")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c, err := startStdio(spec)
	if err != nil {
		return nil, err
	}
	defer c.close()

	raw, err := c.call(ctx, "initialize", map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "raaya", "version": ClientVersion},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize: %w%s", err, c.stderrTail())
	}
	var hello struct {
		Capabilities map[string]json.RawMessage `json:"capabilities"`
		ServerInfo   struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}
	if err := json.Unmarshal(raw, &hello); err != nil {
		return nil, fmt.Errorf("initialize: unexpected result: %w", err)
	}
	if err := c.notify("notifications/initialized"); err != nil {
		return nil, err
	}

	res := &LiveResult{}
	if hello.ServerInfo.Name != "" {
		res.ServerInfo = strings.TrimSuffix(hello.ServerInfo.Name+"@"+hello.ServerInfo.Version, "@")
	}

	if _, ok := hello.Capabilities["tools"]; ok {
		err := c.paginate(ctx, "tools/list", func(page json.RawMessage) (string, error) {
			var p struct {
				Tools      []LiveTool `json:"tools"`
			res.Tools = append(res.Tools, p.Tools...)
			return p.NextCursor, nil
		})
		if err != nil {
			return nil, fmt.Errorf("tools/list: %w", err)
		}
	}

	if _, ok := hello.Capabilities["resources"]; ok {
		// Resources are informational; a failure here is not fatal.
		_ = c.paginate(ctx, "resources/list", func(page json.RawMessage) (string, error) {
			var p struct {
				Resources  []LiveResource `json:"resources"`
				NextCursor string         `json:"nextCursor"`
			}
			if err := json.Unmarshal(page, &p); err != nil {
				return "", err
			}
			res.Resources = append(res.Resources, p.Resources...)
			return p.NextCursor, nil
		})
	}

	sort.Slice(res.Tools, func(i, j int) bool { return res.Tools[i].Name < res.Tools[j].Name })
	sort.Slice(res.Resources, func(i, j int) bool { return res.Resources[i].URI < res.Resources[j].URI })
	return res, nil
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcIn struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcOut struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  any             `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type stdioClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	msgs   chan rpcIn
	done   chan struct{}
	stderr *tailBuffer
	nextID int
}

func startStdio(spec ServerSpec) (*stdioClient, error) {
	args := make([]string, len(spec.Args))
	for i, a := range spec.Args {
		args[i] = expandVars(a, spec.Cwd)
	}
	cmd := exec.Command(spec.Command, args...)
	cmd.Dir = spec.Cwd
	cmd.Env = os.Environ()
	for _, k := range sortedKeys(spec.Env) {
		cmd.Env = append(cmd.Env, k+"="+expandVars(spec.Env[k], spec.Cwd))
	}
	cmd.WaitDelay = 2 * time.Second
	setProcessGroup(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	tail := &tailBuffer{limit: 4096}
	cmd.Stderr = tail
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", spec.Command, err)
	}

	c := &stdioClient{cmd: cmd, stdin: stdin, msgs: make(chan rpcIn, 16), done: make(chan struct{}), stderr: tail}
	go c.readLoop(stdout)
	return c, nil
}

// readLoop forwards newline-delimited JSON-RPC messages, skipping anything
// else a server prints to stdout (log lines, banners).
func (c *stdioClient) readLoop(r io.Reader) {
	defer close(c.msgs)
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadBytes('\n')
		if trimmed := bytes.TrimSpace(line); len(trimmed) > 0 && trimmed[0] == '{' {
			var m rpcIn
			if json.Unmarshal(trimmed, &m) == nil {
				select {
				case c.msgs <- m:
				case <-c.done:
					return
				}
			}
		}
		if err != nil {
			return
		}
	}
}

func (c *stdioClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.nextID++
	id := strconv.Itoa(c.nextID)
	if err := c.send(rpcOut{JSONRPC: "2.0", ID: json.RawMessage(id), Method: method, Params: params}); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case m, ok := <-c.msgs:
			if !ok {
				return nil, fmt.Errorf("server exited%s", c.stderrTail())
			}
			if m.Method != "" { // a request or notification from the server
				if len(m.ID) > 0 {
					c.answerServer(m)
				}
				continue
			}
			if string(bytes.TrimSpace(m.ID)) != id {
				continue
			}
			if m.Error != nil {
				return nil, fmt.Errorf("%s (code %d)", m.Error.Message, m.Error.Code)
			}
			return m.Result, nil
		}
	}
}

func (c *stdioClient) paginate(ctx context.Context, method string, handle func(json.RawMessage) (string, error)) error {
	cursor := ""
	for page := 0; page < maxListPages; page++ {
		params := map[string]any{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		raw, err := c.call(ctx, method, params)
		if err != nil {
			return err
		}
		next, err := handle(raw)
		if err != nil {
			return err
		}
		if next == "" || next == cursor {
			return nil
		}
		cursor = next
	}
	return nil
}

func (c *stdioClient) notify(method string) error {
	return c.send(rpcOut{JSONRPC: "2.0", Method: method})
}

// answerServer replies to server-initiated requests so servers don't hang.
func (c *stdioClient) answerServer(m rpcIn) {
	out := rpcOut{JSONRPC: "2.0", ID: m.ID}
	if m.Method == "ping" {
		out.Result = map[string]any{}
	} else {
		out.Error = &rpcError{Code: -32601, Message: "method not supported by raaya"}
	}
	_ = c.send(out)
}

	}
	_, err = c.stdin.Write(append(data, '\n'))
	return err
}

func (c *stdioClient) close() {
	close(c.done)
	_ = c.stdin.Close()
	killProcess(c.cmd)
	_ = c.cmd.Wait()
}

func (c *stdioClient) stderrTail() string {
	if s := strings.TrimSpace(c.stderr.String()); s != "" {
		return "; stderr: " + s
	}
	return ""
}

// expandVars resolves $VAR and ${VAR} from the environment, plus VS Code's
// ${workspaceFolder}. Unknown references (e.g. ${input:token}) become empty.
func expandVars(s, workspace string) string {
	return os.Expand(s, func(name string) string {
		if name == "workspaceFolder" && workspace != "" {
			return workspace
		}
		if v, ok := strings.CutPrefix(name, "env:"); ok {
			return os.Getenv(v)
		}
		return os.Getenv(name)
	})
}

// tailBuffer keeps the last limit bytes written to it.
type tailBuffer struct {
	mu    sync.Mutex
	buf   []byte
	limit int
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, p...)
	if len(t.buf) > t.limit {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-t.limit:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.buf)
}
