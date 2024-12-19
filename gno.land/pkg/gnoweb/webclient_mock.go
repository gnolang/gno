package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// MockPackage represents a mock package with files and function signatures.
type MockPackage struct {
	Path      string
	Domain    string
	Files     map[string] /* filename  */ string /* body */
	Functions []vm.FunctionSignature
}

// MockWebClient is a mock implementation of the Client interface.
type MockWebClient struct {
	Packages map[string] /* path */ *MockPackage /* package */
}

func NewMockWebClient(pkgs ...*MockPackage) *MockWebClient {
	mpkgs := make(map[string]*MockPackage)
	for _, pkg := range pkgs {
		mpkgs[pkg.Path] = pkg
	}

	return &MockWebClient{Packages: mpkgs}
}

// Render simulates rendering a package by writing its content to the writer.
func (m *MockWebClient) RenderRealm(w io.Writer, path string, args string) (*RealmMeta, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	fmt.Fprintf(w, "[%s]%s:", pkg.Domain, pkg.Path)

	// Return a dummy RealmMeta for simplicity
	return &RealmMeta{}, nil
}

// SourceFile simulates retrieving a source file's metadata.
func (m *MockWebClient) SourceFile(w io.Writer, pkgPath, fileName string) (*FileMeta, error) {
	pkg, exists := m.Packages[pkgPath]
	if !exists {
		return nil, errors.New("package not found")
	}

	if body, ok := pkg.Files[fileName]; ok {
		w.Write([]byte(body))
		return &FileMeta{
			Lines:  len(bytes.Split([]byte(body), []byte("\n"))),
			SizeKb: float64(len(body)) / 1024.0,
		}, nil
	}

	return nil, errors.New("file not found")
}

// Functions simulates retrieving function signatures from a package.
func (m *MockWebClient) Functions(path string) ([]vm.FunctionSignature, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, errors.New("package not found")
	}

	return pkg.Functions, nil
}

// Sources simulates listing all source files in a package.
func (m *MockWebClient) Sources(path string) ([]string, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, errors.New("package not found")
	}

	fileNames := make([]string, 0, len(pkg.Files))
	for file, _ := range pkg.Files {
		fileNames = append(fileNames, file)
	}

	// Sort for consistency
	sort.Strings(fileNames)

	return fileNames, nil
}
