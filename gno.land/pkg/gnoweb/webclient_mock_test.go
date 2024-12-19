package gnoweb_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// MockPackage represents a mock package with files and function signatures.
type MockPackage struct {
	Domain    string
	Path      string
	Files     map[string] /* filename  */ string /* body */
	Functions []vm.FunctionSignature
}

// MockClient is a mock implementation of the gnoweb.Client interface.
type MockClient struct {
	Packages map[string] /* path */ *MockPackage /* package */
}

// Render simulates rendering a package by writing its content to the writer.
func (m *MockClient) RenderRealm((w io.Writer, path string, args string) (*gnoweb.RealmMeta, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, errors.New("package not found")
	}

	fmt.Fprintf(w, "<code>[%s]%s:%s</code>", pkg.Domain, pkg.Path)

	// Return a dummy RealmMeta for simplicity
	return &gnoweb.RealmMeta{}, nil
}

// SourceFile simulates retrieving a source file's metadata.
func (m *MockClient) SourceFile(w io.Writer, pkgPath, fileName string) (*gnoweb.FileMeta, error) {
	pkg, exists := m.Packages[pkgPath]
	if !exists {
		return nil, errors.New("package not found")
	}

	if body, ok := pkg.Files[fileName]; ok {
		w.Write([]byte(body))
		return &gnoweb.FileMeta{
			Lines:  len(bytes.Split([]byte(body), []byte("\n"))),
			SizeKb: float64(len(body)) / 1024.0,
		}, nil
	}

	return nil, errors.New("file not found")
}

// Functions simulates retrieving function signatures from a package.
func (m *MockClient) Functions(path string) ([]vm.FunctionSignature, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, errors.New("package not found")
	}

	return pkg.Functions, nil
}

// Sources simulates listing all source files in a package.
func (m *MockClient) Sources(path string) ([]string, error) {
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
