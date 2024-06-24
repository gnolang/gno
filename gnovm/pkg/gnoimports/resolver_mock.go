package gnoimports

import (
	"bytes"
	"fmt"
	"io"
)

type mockResolver struct {
	pkgspath map[string]Package   // pkg path -> pkg
	pkgs     map[string][]Package // pkg name -> []pkg
}

func newMockResolver() *mockResolver {
	return &mockResolver{
		pkgspath: make(map[string]Package),
		pkgs:     make(map[string][]Package),
	}
}

type mockPackage struct {
	PkgPath   string
	PkgName   string
	filesname []string
	files     [][]byte
}

func newMockedPackage(path, name string) *mockPackage {
	return &mockPackage{PkgPath: path, PkgName: name}
}

func (m *mockPackage) AddFile(filename string, body []byte) {
	m.filesname = append(m.filesname, filename)
	m.files = append(m.files, body)
}

// Should return the package path
func (m *mockPackage) Path() string {
	return m.PkgPath
}

// Should return the name of the as definied at the top level of each
// files
func (m *mockPackage) Name() string {
	return m.PkgName
}

// Should return all gno filename inside the package
func (m *mockPackage) Files() []string {
	return m.filesname
}

// ReaderCloser wraps an io.Reader and provides a no-op Close method.
type readerCloser struct {
	io.Reader
}

func (readerCloser) Close() error { return nil }

// Should return a content reader for the the given filename within the package
func (m *mockPackage) Read(filename string) (io.ReadCloser, error) {
	for i, file := range m.filesname {
		if file != filename {
			continue
		}

		r := bytes.NewReader(m.files[i])
		return &readerCloser{r}, nil
	}

	return nil, fmt.Errorf("file not found %q", filename)
}

func (m *mockResolver) AddPackage(pkg Package) []Package {
	m.pkgs[pkg.Name()] = append(m.pkgs[pkg.Name()], pkg)
	m.pkgspath[pkg.Path()] = pkg
	return nil
}

func (m *mockResolver) ResolveName(pkgname string) []Package {
	return m.pkgs[pkgname]
}

func (m *mockResolver) ResolvePath(pkgpath string) Package {
	return m.pkgspath[pkgpath]
}
