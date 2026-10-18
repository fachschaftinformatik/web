package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/fachschaftinformatik/web"

// Check transitive local dependencies as well as direct imports so a shared
// helper cannot accidentally reconnect layers. Tests obey the same boundaries.
func TestLayerDependencies(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../..")
	graph := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		pkg := module + "/" + filepath.ToSlash(relative)
		for _, spec := range file.Imports {
			dependency, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			graph[pkg] = append(graph[pkg], dependency)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(graph[module+"/view"]) == 0 || len(graph[module+"/site"]) == 0 {
		t.Fatal("architecture check did not discover the application packages")
	}
	for pkg := range graph {
		var forbidden []string
		switch {
		case within(pkg, module+"/model"):
			forbidden = []string{module + "/site", module + "/view"}
		case within(pkg, module+"/view"):
			forbidden = []string{module + "/model", module + "/site", module + "/storage", module + "/internal/database", "database/sql"}
		case within(pkg, module+"/storage"):
			forbidden = []string{module + "/model", module + "/site", module + "/view"}
		default:
			continue
		}
		if path := forbiddenPath(graph, pkg, forbidden, map[string]bool{}); len(path) > 0 {
			t.Errorf("forbidden layer dependency: %s", strings.Join(path, " → "))
		}
	}
}

func within(path, prefix string) bool { return path == prefix || strings.HasPrefix(path, prefix+"/") }

func forbiddenPath(graph map[string][]string, pkg string, forbidden []string, visited map[string]bool) []string {
	if visited[pkg] {
		return nil
	}
	visited[pkg] = true
	for _, dependency := range graph[pkg] {
		for _, prefix := range forbidden {
			if within(dependency, prefix) {
				return []string{pkg, dependency}
			}
		}
		if path := forbiddenPath(graph, dependency, forbidden, visited); len(path) > 0 {
			return append([]string{pkg}, path...)
		}
	}
	return nil
}

func TestForbiddenPathFindsTransitiveLayerViolations(t *testing.T) {
	graph := map[string][]string{
		"view":    {"shared", "modelish"},
		"shared":  {"utility"},
		"utility": {"shared", "model/records"},
	}
	path := forbiddenPath(graph, "view", []string{"model"}, map[string]bool{})
	if strings.Join(path, " → ") != "view → shared → utility → model/records" {
		t.Fatalf("indirect model dependency not reported: %v", path)
	}
	if path := forbiddenPath(graph, "view", []string{"storage", "modelish/subpackage"}, map[string]bool{}); len(path) != 0 {
		t.Fatalf("unrelated package or import cycle rejected: %v", path)
	}
}
