package gnolang_test

import (
	"io"
	"math"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
)

// GetNameSourceForPath must tolerate package-block names that have no
// declaring file in the FileSet. The REPL defines earlier-statement names
// directly in the package block of a synthetic package whose FileSet is
// empty, so resolving such a path must return a nil *FileNode rather than
// panic -- a path-keyed check on assignment targets is the first caller
// that runs on those synthetic packages.
func TestGetNameSourceForPathNoFilesetDecl(t *testing.T) {
	_, store := test.TestStore(gnoenv.RootDir(), io.Discard, nil)

	pn := gno.NewPackageNode("repltest", "gno.land/r/repltest", &gno.FileSet{})
	pv := pn.NewPackage(gno.NewAllocator(math.MaxInt64))
	store.SetBlockNode(pn)
	store.SetCachePackage(pv)

	// Define a name directly in the package block, with no declaring decl
	// in the (empty) FileSet -- the shape the REPL produces for statements
	// that earlier input defined, e.g. `var outVar string = "x"`.
	vd := &gno.ValueDecl{}
	nx := &gno.NameExpr{Name: "outVar"}
	pn.Reserve(false, nx, vd, gno.NSValueDecl, 0)

	idx, ok := pn.GetLocalIndex("outVar")
	if !ok {
		t.Fatal("name not defined in package block")
	}
	path := gno.ValuePath{
		Type: gno.VPBlock,
		Name: "outVar",
	}
	path.SetDepth(1)
	path.Index = idx

	_, fn, nsrc := pn.GetNameSourceForPath(store, path)
	if fn != nil {
		t.Fatalf("expected nil FileNode for synthetic package, got %v", fn.FileName)
	}
	if nsrc.Type != gno.NSValueDecl {
		t.Fatalf("expected NSValueDecl name source, got %v", nsrc.Type)
	}
}
