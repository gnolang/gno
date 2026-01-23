package gnoweb

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// MockPackage represents a mock package with files and function signatures for testing.
type MockPackage struct {
	Path      string
	Domain    string
	Files     map[string]string // filename -> body
	Functions []*doc.JSONFunc
}

// MockClient is a mock implementation of the ClientAdapter interface for testing.
type MockClient struct {
	Packages map[string]*MockPackage // path -> package
}

var _ ClientAdapter = (*MockClient)(nil)

// NewMockClient creates a new MockClient from one or more MockPackages.
func NewMockClient(pkgs ...*MockPackage) *MockClient {
	mpkgs := make(map[string]*MockPackage)
	for _, pkg := range pkgs {
		mpkgs[pkg.Path] = pkg
	}
	return &MockClient{Packages: mpkgs}
}

// Realm fetches the content of a realm from a given path and returns the data, or an error if not found or not declared.
func (m *MockClient) Realm(ctx context.Context, path, args string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPackageNotFound
	}
	if !pkgHasRender(pkg) {
		return nil, ErrClientRenderNotDeclared
	}
	// Simulate output: [domain]/path:args
	header := fmt.Sprintf("# [%s]/%s:%s\n\n", pkg.Domain, strings.Trim(path, "/"), args)
	var body string
	for name, content := range pkg.Files {
		body += fmt.Sprintf("# %s\n```\n%s\n```\n\n", name, content)
	}

	return []byte(header + body), nil
}

// File fetches the source file from a given package path and filename, returning its content and metadata.
func (m *MockClient) File(ctx context.Context, pkgPath, fileName string) ([]byte, FileMeta, error) {
	if err := ctx.Err(); err != nil {
		return nil, FileMeta{}, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[pkgPath]
	if !exists {
		return nil, FileMeta{}, ErrClientPackageNotFound
	}
	body, ok := pkg.Files[fileName]
	if !ok {
		return nil, FileMeta{}, ErrClientPackageNotFound
	}
	// Calculate metadata
	lines := len(bytes.Split([]byte(body), []byte("\n")))
	sizeKb := float64(len(body)) / 1024.0
	meta := FileMeta{
		Lines:  lines,
		SizeKB: sizeKb,
	}
	return []byte(body), meta, nil
}

// ListFiles lists all source files available in a specified package path.
func (m *MockClient) ListFiles(ctx context.Context, path string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPackageNotFound
	}
	fileNames := make([]string, 0, len(pkg.Files))
	for file := range pkg.Files {
		fileNames = append(fileNames, file)
	}
	sort.Strings(fileNames)
	return fileNames, nil
}

// ListPaths lists all package paths that match the specified prefix, up to the given limit.
func (m *MockClient) ListPaths(ctx context.Context, prefix string, limit int) ([]string, error) {
	var shouldKeep func(s string) bool
	if strings.HasPrefix(prefix, "@") {
		name := prefix[1:]
		shouldKeep = func(s string) bool {
			return strings.HasPrefix(s, "/r/"+name) ||
				strings.HasPrefix(s, "/p/"+name)
		}
	} else {
		shouldKeep = func(s string) bool {
			return strings.HasPrefix(s, prefix)
		}
	}
	list := []string{}
	for _, pkg := range m.Packages {
		if len(list) >= limit {
			break
		}
		if shouldKeep(pkg.Path) {
			list = append(list, pkg.Path)
		}
	}
	return list, nil
}

// Doc retrieves the JSON documentation for a specified package path.
func (m *MockClient) Doc(ctx context.Context, path string) (*doc.JSONDocumentation, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPackageNotFound
	}
	return &doc.JSONDocumentation{Funcs: pkg.Functions}, nil
}

// Helper: check if package has a Render(string) string function.
func pkgHasRender(pkg *MockPackage) bool {
	if len(pkg.Functions) == 0 {
		return false
	}
	for _, fn := range pkg.Functions {
		if fn.Name == "Render" &&
			len(fn.Params) == 1 &&
			len(fn.Results) == 1 &&
			fn.Params[0].Type == "string" &&
			fn.Results[0].Type == "string" {
			return true
		}
	}
	return false
}
