package gnoweb

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// MockPackage represents a mock package with files and function signatures for testing.
type MockPackage struct {
	Path      string
	Domain    string
	Files     map[string]string // filename -> body
	Functions []*doc.JSONFunc
	// Inert stages a package that was submitted but not yet approved. The
	// other methods still refuse it the way the chain does -- Render and the
	// file queries read the live key space -- so a test gets the real shape.
	Inert bool
	// Pending stages a redeploy parked over a live package.
	Pending bool
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

// PackageMeta reports the mock's view of a path. A package present in the mock
// is live; Pending is set from MockPackage.Pending so a test can stage the
// parked case, which is the one gnoweb has to render differently.
func (m *MockClient) PackageMeta(ctx context.Context, path string) (*vm.PackageMeta, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// rpcClient trims before querying; mirror it so a caller that passes a
	// directory-style path gets the same answer from both.
	path = strings.TrimSuffix(path, "/")
	pkg, exists := m.Packages[path]
	if !exists {
		return &vm.PackageMeta{Path: path, Status: vm.PackageStatusAbsent}, nil
	}
	meta := &vm.PackageMeta{Path: path, Status: vm.PackageStatusLive, Pending: pkg.Pending}
	if pkg.Inert {
		meta.Status = vm.PackageStatusInert
		meta.Pending = true
	}
	return meta, nil
}

// Realm fetches the content of a realm from a given path and returns the data, or an error if not found or not declared.
func (m *MockClient) Realm(ctx context.Context, path, args string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[path]
	if !exists || pkg.Inert {
		// Inert packages live in a separate key space that this query cannot
		// read, so the chain answers not-found for them too.
		return nil, ErrClientPackageNotFound
	}
	if !pkgHasRender(pkg) {
		return nil, ErrClientRenderNotDeclared
	}
	// Simulate output: [domain]/path:args
	header := fmt.Sprintf("# [%s]/%s:%s\n\n", pkg.Domain, strings.Trim(path, "/"), args)
	var body strings.Builder
	for name, content := range pkg.Files {
		body.WriteString(fmt.Sprintf("# %s\n```\n%s\n```\n\n", name, content))
	}

	return []byte(header + body.String()), nil
}

// File fetches the source file from a given package path and filename, returning its content and metadata.
func (m *MockClient) File(ctx context.Context, pkgPath, fileName string, _ int64) ([]byte, FileMeta, error) {
	if err := ctx.Err(); err != nil {
		return nil, FileMeta{}, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[pkgPath]
	if !exists || pkg.Inert {
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
func (m *MockClient) ListFiles(ctx context.Context, path string, _ int64) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[path]
	if !exists || pkg.Inert {
		// Inert packages live in a separate key space that this query cannot
		// read, so the chain answers not-found for them too.
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
		// Parked packages are stored under a separate key prefix that
		// FindPathsByPrefix does not range over, so the chain never lists them.
		if !pkg.Inert && shouldKeep(pkg.Path) {
			list = append(list, pkg.Path)
		}
	}
	return list, nil
}

// Doc retrieves the JSON documentation for a specified package path.
func (m *MockClient) Doc(ctx context.Context, path string, _ int64) (*doc.JSONDocumentation, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	pkg, exists := m.Packages[path]
	if !exists || pkg.Inert {
		// Inert packages live in a separate key space that this query cannot
		// read, so the chain answers not-found for them too.
		return nil, ErrClientPackageNotFound
	}
	return &doc.JSONDocumentation{Funcs: pkg.Functions}, nil
}

// StatePkg returns mock package state data for testing.
func (m *MockClient) StatePkg(_ context.Context, path string, _ int64) ([]byte, error) {
	pkg, exists := m.Packages[path]
	if !exists || pkg.Inert {
		return nil, ErrClientPackageNotFound
	}
	// Empty package shape matching what the keeper produces via amino.MarshalJSON.
	return []byte(`{"names":[],"values":[]}`), nil
}

// StateObject returns mock object state data for testing.
func (m *MockClient) StateObject(_ context.Context, _ string, _ int64) ([]byte, error) {
	// Empty StructValue — minimal valid shape for amino.UnmarshalJSON.
	return []byte(`{"objectid":"","value":{"@type":"/gno.StructValue","Fields":[]}}`), nil
}

// StateType returns mock type data for testing.
func (m *MockClient) StateType(_ context.Context, _ string, _ int64) ([]byte, error) {
	return []byte(`{"typeid":"","type":{"@type":"/gno.PrimitiveType","value":"32"}}`), nil
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
