package gnoweb

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"sort"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// MockPackage represents a mock package with files and function signatures.
type MockPackage struct {
	Path      string
	Domain    string
	Files     map[string]string // filename -> body
	Functions []*doc.JSONFunc
}

// MockWebClient is a mock implementation of the Client interface.
type MockWebClient struct {
	Packages map[string]*MockPackage // path -> package
}

var _ WebClient = (*MockWebClient)(nil)

func NewMockWebClient(pkgs ...*MockPackage) *MockWebClient {
	mpkgs := make(map[string]*MockPackage)
	for _, pkg := range pkgs {
		mpkgs[pkg.Path] = pkg
	}

	return &MockWebClient{Packages: mpkgs}
}

// RenderRealm simulates rendering a package by writing its content to the writer.
func (m *MockWebClient) RenderRealm(w io.Writer, u *weburl.GnoURL, _ ContentRenderer) (*RealmMeta, error) {
	pkg, exists := m.Packages[u.Path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	if !pkgHasRender(pkg) {
		return nil, ErrRenderNotDeclared
	}

	// Return the production format [domain]/path:args
	fmt.Fprintf(w, "[%s]/%s:%s", pkg.Domain, strings.Trim(u.Path, "/"), u.Args)

	// Return a dummy RealmMeta for simplicity
	return &RealmMeta{}, nil
}

// SourceFile simulates retrieving a source file's metadata.
func (m *MockWebClient) SourceFile(w io.Writer, pkgPath, fileName string, isRaw bool) (*FileMeta, error) {
	pkg, exists := m.Packages[pkgPath]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	if body, ok := pkg.Files[fileName]; ok {
		w.Write([]byte(body))
		return &FileMeta{
			Lines:  len(bytes.Split([]byte(body), []byte("\n"))),
			SizeKb: float64(len(body)) / 1024.0,
		}, nil
	}

	return nil, ErrClientPathNotFound
}

// Doc simulates retrieving function docs from a package.
func (m *MockWebClient) Doc(path string) (*doc.JSONDocumentation, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	return &doc.JSONDocumentation{Funcs: pkg.Functions}, nil
}

// Sources simulates listing all source files in a package.
func (m *MockWebClient) Sources(path string) ([]string, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	fileNames := make([]string, 0, len(pkg.Files))
	for file := range pkg.Files {
		fileNames = append(fileNames, file)
	}

	// Sort for consistency
	sort.Strings(fileNames)

	return fileNames, nil
}

func (m *MockWebClient) iterPath(filter func(s string) bool) iter.Seq[string] {
	return func(yield func(v string) bool) {
		for _, pkg := range m.Packages {
			if filter(pkg.Path) {
				continue
			}

			if !yield(pkg.Path) {
				return
			}
		}
	}
}

// Sources simulates listing all source files in a package.
func (m *MockWebClient) QueryPaths(prefix string, limit int) ([]string, error) {
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
	m.iterPath(func(s string) bool {
		if len(list) > limit {
			return false
		}

		if shouldKeep(s) {
			list = append(list, s)
		}

		return true
	})
	return list, nil
}

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
