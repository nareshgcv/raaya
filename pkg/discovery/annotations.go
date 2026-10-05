package discovery

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// SourceTool is an MCP tool definition (or an @raaya:capability-annotated
// function) found in source code.
type SourceTool struct {
	Name        string
	File        string // slash-separated, relative to the scan root
	Line        int
	Permissions []graph.PermissionLevel // from @raaya:capability; empty if none
}

var (
	capabilityPattern = regexp.MustCompile(`@raaya:capability\s+([A-Za-z][A-Za-z ,|]*)`)
	pyToolDecorator   = regexp.MustCompile(`^@(?:[\w.]+\.)?tool\b(?:\((.*))?`)
	pyNameArg         = regexp.MustCompile(`name\s*=\s*["']([^"']+)["']`)
	pyDef             = regexp.MustCompile(`^(?:async\s+)?def\s+(\w+)\s*\(`)
	jsTool            = regexp.MustCompile("\\.(?:tool|registerTool)\\s*\\(\\s*[\"'`]([^\"'`]+)[\"'`]")
	jsFunc            = regexp.MustCompile(`(?:^|\s)function\s+(\w+)`)
	goNewTool         = regexp.MustCompile(`NewTool\(\s*"([^"]+)"`)
	goToolStruct      = regexp.MustCompile(`Tool\{\s*Name:\s*"([^"]+)"`)
	goFunc            = regexp.MustCompile(`^func\s+(?:\([^)]*\)\s*)?(\w+)\s*\(`)
)

var skipDirs = map[string]bool{
	"node_modules": true, "vendor": true, "dist": true, "build": true, "target": true,
	"venv": true, "__pycache__": true, "testdata": true,
}

var sourceLangs = map[string]string{
func ScanSource(root string) ([]SourceTool, error) {
	var tools []SourceTool
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			return nil // unreadable entry: skip it
		}
		if d.IsDir() {
			if path != root && (skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		lang, ok := sourceLangs[filepath.Ext(path)]
		if !ok {
			return nil
		}
		if info, err := d.Info(); err != nil || info.Size() > maxSourceFileSize {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		if found, err := scanFile(path, filepath.ToSlash(rel), lang); err == nil {
			tools = append(tools, found...)
		}
		return nil
	})
	sort.SliceStable(tools, func(i, j int) bool {
		if tools[i].File != tools[j].File {
			return tools[i].File < tools[j].File
		}
		return tools[i].Line < tools[j].Line
	})
	return tools, err
}

func scanFile(path, rel, lang string) ([]SourceTool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), maxSourceFileSize)

	var (
		out      []SourceTool
		caps     []graph.PermissionLevel
		capsLine int
		lineNo   int
		pyDecor  bool
		pyName   string
	)
	emit := func(name string) {
		t := SourceTool{Name: name, File: rel, Line: lineNo}
		if caps != nil && lineNo-capsLine <= annotationReach {
			t.Permissions = caps
		}
		out = append(out, t)
		caps = nil
	}

	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if m := capabilityPattern.FindStringSubmatch(line); m != nil {
			caps, capsLine = parseCapabilities(m[1]), lineNo
			continue
		}
		pending := caps != nil && lineNo-capsLine <= annotationReach

		switch lang {
		case "py":
			if m := pyToolDecorator.FindStringSubmatch(line); m != nil {
				pyDecor, pyName = true, ""
				if n := pyNameArg.FindStringSubmatch(m[1]); n != nil {
					pyName = n[1]
				}
				continue
			}
			if m := pyDef.FindStringSubmatch(line); m != nil {
				switch {
				case pyDecor && pyName != "":
					emit(pyName)
				case pyDecor || pending:
					emit(m[1])
				}
				pyDecor, pyName = false, ""
			}
		case "js":
			if m := jsTool.FindStringSubmatch(line); m != nil {
				emit(m[1])
			} else if m := jsFunc.FindStringSubmatch(line); m != nil && pending {
				emit(m[1])
			}
		case "go":
			if m := goNewTool.FindStringSubmatch(line); m != nil {
				emit(m[1])
			} else if m := goToolStruct.FindStringSubmatch(line); m != nil {
				emit(m[1])
			} else if m := goFunc.FindStringSubmatch(line); m != nil && pending {
				emit(m[1])
			}
		}
	}
	return out, sc.Err()
}

func parseCapabilities(s string) []graph.PermissionLevel {
	var out []graph.PermissionLevel
	seen := map[graph.PermissionLevel]bool{}
	for _, field := range strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == '|' }) {
		if p, ok := graph.ParsePermission(field); ok && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
