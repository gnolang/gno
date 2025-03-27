package gnodebug

import (
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

const (
	gnolang  = "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnodebug = gnolang + "/gnodebug"
)

func TestDocFlags(t *testing.T) {
	// Do static analysis on pkg/gnolang to find uses of DebugType methods.
	// Ensure that the types that they use are registered in the docs.
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.LoadSyntax,
		},
		gnolang, gnodebug,
	)
	require.NoError(t, err)
	require.Len(t, pkgs, 2)
	gnop := slices.IndexFunc(pkgs, func(pkg *packages.Package) bool { return pkg.PkgPath == gnolang })
	gno := pkgs[gnop]
	names := findFlagNames(gno)
	for _, fd := range FlagDocs {
		pos := slices.Index(names, fd.Name)
		if pos < 0 {
			t.Errorf("flag is documented but not used: %q", fd.Name)
		} else {
			names = slices.Delete(names, pos, pos+1)
		}
	}
	if len(names) > 0 {
		t.Errorf("flag(s) are not documented: %v", names)
	}
}

// finds all uses of DebugType.{Printf,Enabled,Get,Set}, and if the first
// argument is a constant string, add it to a set of names to return.
func findFlagNames(pkg *packages.Package) []string {
	names := make([]string, 0, 32)
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			cx, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sx, ok := cx.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			tp, ok := pkg.TypesInfo.TypeOf(sx.X).(*types.Named)
			if !ok {
				return true
			}
			tpPkg := tp.Obj().Pkg()
			if tpPkg == nil || tpPkg.Path() != gnodebug {
				return true
			}
			if tp.Obj().Name() != "DebugType" {
				return true
			}
			switch sx.Sel.Name {
			case "Printf", "Enabled", "Get", "Set":
				tv := pkg.TypesInfo.Types[cx.Args[0]]
				if tv.Value == nil {
					return true
				}
				sv := constant.StringVal(tv.Value)
				if sv == "" {
					return true
				}
				if !slices.Contains(names, sv) {
					names = append(names, sv)
				}
			}
			return true
		})
	}
	return names
}
