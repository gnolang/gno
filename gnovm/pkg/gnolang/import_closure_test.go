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
