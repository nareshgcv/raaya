package discovery

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// TestMain lets the test binary double as a fake MCP server, so the live
// client is tested against a real child process without network or Node.
func TestMain(m *testing.M) {
	if os.Getenv("RAAYA_FAKE_MCP") == "1" {
		runFakeServer()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func runFakeServer() {
	fmt.Println("fake server booting") // non-JSON noise the client must skip
	in := bufio.NewScanner(os.Stdin)
	out := json.NewEncoder(os.Stdout)
	reply := func(id json.RawMessage, result any) {
		_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
	}
	for in.Scan() {
		var req struct {
			ID     json.RawMessage `json:"id"`
			reply(req.ID, map[string]any{
				"protocolVersion": mcpProtocolVersion,
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": "fake", "version": "1.0.0"},
			})
		case "tools/list":
			if req.Params.Cursor == "" {
				// A server-initiated request mid-flight; the client must answer it.
				_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": "srv-1", "method": "ping"})
				reply(req.ID, map[string]any{
					"tools":      []any{map[string]any{"name": "read_file", "annotations": map[string]any{"readOnlyHint": true}}},
					"nextCursor": "page-2",
				})
			} else {
				reply(req.ID, map[string]any{"tools": []any{map[string]any{"name": "run_shell"}}})
			}
		default:
			_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -32601, "message": "nope"}})
		}
	}
}

func TestListServerAgainstFakeServer(t *testing.T) {
	spec := ServerSpec{
		Name:    "fake",
		Command: os.Args[0],
		Args:    []string{"-test.run=^$"},
		Env:     map[string]string{"RAAYA_FAKE_MCP": "1"},
	}
	res, err := ListServer(context.Background(), spec, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if res.ServerInfo != "fake@1.0.0" {
		t.Errorf("server info: %q", res.ServerInfo)
	}
	if len(res.Tools) != 2 {
		t.Fatalf("want 2 tools across both pages, got %+v", res.Tools)
	}
	want := map[string]graph.PermissionLevel{"read_file": graph.PermRead, "run_shell": graph.PermExecute}
	for _, tool := range res.Tools {
		if got, _ := InferToolPermission(tool.Name, tool.Annotations); got != want[tool.Name] {
			t.Errorf("%s: got %s, want %s", tool.Name, got, want[tool.Name])
		}
	}
}

func TestListServerReportsStartFailure(t *testing.T) {
	_, err := ListServer(context.Background(), ServerSpec{Name: "missing", Command: "raaya-no-such-binary"}, 5*time.Second)
	if err == nil {
		t.Fatal("expected an error for a missing command")
	}
}
