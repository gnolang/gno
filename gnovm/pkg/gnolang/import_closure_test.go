package gnolang

import (
	"strings"
	"testing"
)

// closureTestStore answers only the lookups importClosure makes.
type closureTestStore struct {
	Store
	nodes map[string]*PackageNode
}

func (s *closureTestStore) GetBlockNodeSafe(loc Location) BlockNode {
	if pn := s.nodes[loc.PkgPath]; pn != nil {
		return pn
	}
	return nil
}

func importDecl(path string) *ImportDecl {
	name := path
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	return &ImportDecl{NameExpr: NameExpr{Name: Name(name)}, PkgPath: path}
}

func filesFor(imports ...string) []*FileNode {
	decls := make(Decls, len(imports))
	for i, p := range imports {
		decls[i] = importDecl(p)
	}
	return []*FileNode{{Decls: decls}}
}

// newClosureFixture is a realm importing 10 stdlib and 10 /p/ packages, each
// with imports of its own.
func newClosureFixture() (*closureTestStore, []*FileNode) {
	std := []string{
		"fmt", "sort", "strings", "bytes", "bufio",
		"io", "regexp", "strconv", "time", "unicode",
	}
	libs := make([]string, 10)
	for i := range libs {
		libs[i] = "gno.land/p/bench/pkg" + string(rune('a'+i))
	}
	store := &closureTestStore{nodes: map[string]*PackageNode{}}
	for i, path := range std {
		var imps []string
		if i > 0 {
			imps = append(imps, std[i-1])
		}
		store.nodes[path] = &PackageNode{PkgPath: path, FileSet: &FileSet{Files: filesFor(imps...)}}
	}
	for i, path := range libs {
		imps := []string{std[i]}
		if i > 0 {
			imps = append(imps, libs[i-1])
		}
		store.nodes[path] = &PackageNode{PkgPath: path, FileSet: &FileSet{Files: filesFor(imps...)}}
	}
	realmImports := append(append([]string{}, std...), libs...)
	return store, filesFor(realmImports...)
}

func TestImportClosureSkipsTestFiles(t *testing.T) {
	store := &closureTestStore{nodes: map[string]*PackageNode{
		"bytes": {PkgPath: "bytes", FileSet: &FileSet{Files: filesFor()}},
	}}
	prod := filesFor("io")
	prod[0].FileName = "realm.gno"
	test := filesFor("bytes")
	test[0].FileName = "realm_test.gno"
	filetest := filesFor("testing")
	filetest[0].FileName = "realm_filetest.gno"
	pn := &PackageNode{
		PkgPath: "gno.land/r/x",
		FileSet: &FileSet{Files: append(prod, append(test, filetest...)...)},
	}
	c := pn.importClosure(store)
	if c["bytes"] || c["testing"] {
		t.Fatalf("closure trusts test-only imports: %v", c)
	}
	if !c["io"] {
		t.Fatalf("closure dropped a production import: %v", c)
	}
	it := &PackageNode{
		PkgPath: "gno.land/r/x_test",
		FileSet: &FileSet{Files: filesFor("bytes")},
	}
	if c := it.importClosure(store); len(c) != 0 {
		t.Fatalf("integration test package has a closure: %v", c)
	}
	// Stdlib unit tests are not realm trust roots: their test imports stay
	// trusted, or bytes/stdlib methods would run view-only in TestStdlibs.
	stdProd := filesFor("io")
	stdProd[0].FileName = "hex.gno"
	stdTest := filesFor("bytes")
	stdTest[0].FileName = "hex_test.gno"
	std := &PackageNode{
		PkgPath: "encoding/hex",
		FileSet: &FileSet{Files: append(stdProd, stdTest...)},
	}
	if c := std.importClosure(store); !c["bytes"] {
		t.Fatalf("stdlib test imports were dropped: %v", c)
	}
}

func BenchmarkImportClosure(b *testing.B) {
	store, realmFiles := newClosureFixture()
	newRealm := func() *PackageNode {
		return &PackageNode{
			PkgPath: "gno.land/r/bench/realm",
			FileSet: &FileSet{Files: realmFiles},
		}
	}
	b.Run("cold", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			newRealm().importClosure(store)
		}
	})
	b.Run("cached", func(b *testing.B) {
		pn := newRealm()
		pn.importClosure(store)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pn.importClosure(store)
		}
	})
}
