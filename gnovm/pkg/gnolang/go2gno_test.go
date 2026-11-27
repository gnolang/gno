package gnolang

import (
	"fmt"
	goparser "go/parser"
	"go/token"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseForLoop(t *testing.T) {
	t.Parallel()

	gocode := `package main
func main(){
	for i:=0; i<10; i++ {
		if i == -1 {
			return
		}
	}
}`
	var m *Machine
	n, err := m.ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")
	t.Logf("CODE:\n%s\n\n", gocode)
	t.Logf("AST:\n%#v\n\n", n)
	t.Logf("AST.String():\n%s\n", n.String())
}

func TestParseTilde(t *testing.T) {
	t.Parallel()

	var m *Machine
	_, err := m.ParseFile("tilde.gno", "package A\nvar x = ~0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tilde operator is not permitted")
}

// nestedNameLists returns a struct type nesting depth levels of names fields
// that share one anonymous struct type.
func nestedNameLists(depth, names int) string {
	fields := make([]string, names)
	for i := range fields {
		fields[i] = "f" + strconv.Itoa(i)
	}
	list := strings.Join(fields, ", ")
	typ := "int"
	for range depth {
		typ = "struct{ " + list + " " + typ + " }"
	}
	return typ
}

// tcCopies type checks a one-file package, where the chain applies checkCopies.
func tcCopies(body string) error {
	_, err := TypeCheckMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "grid",
		Path:  "gno.land/r/test/grid",
		Files: []*std.MemFile{{Name: "grid.gno", Body: body}},
	}, TypeCheckOptions{Mode: TCLatestRelaxed, ProdOnly: true})
	return err
}

func TestRejectsExpandingNameLists(t *testing.T) {
	t.Parallel()

	var m *Machine
	// 10^8 copies at depth 8, and 10^64 at depth 64 must not overflow the count.
	// 2 names over 22 to 30 levels stalled predefine for 0.2s to 31.7s.
	// 2 names over 13 levels fit a fixed per-file bound in 241 bytes, which a
	// package repeated across files.
	for _, tc := range []struct{ depth, names int }{
		{8, 10}, {64, 10}, {22, 2}, {26, 2}, {30, 2}, {13, 2},
	} {
		typ := nestedNameLists(tc.depth, tc.names)
		err := tcCopies("package grid\ntype X " + typ)
		require.ErrorContains(t, err, "declarations copy", "file, %+v", tc)
		_, err = m.ParseExpr(typ + "{}")
		require.ErrorContains(t, err, "declarations copy", "expr, %+v", tc)
		_, err = m.ParseStmts("var x " + typ)
		require.ErrorContains(t, err, "declarations copy", "stmts, %+v", tc)
	}
}

// constBlock returns a const block whose first spec is head followed by the
// sum of n ones, and n-1 bare specs that each copy it.
func constBlock(n int, head string) string {
	var sb strings.Builder
	sb.WriteString("const (\n\tc0 " + head + "1" + strings.Repeat("+1", n-1) + "\n")
	for i := 1; i < n; i++ {
		fmt.Fprintf(&sb, "\tc%d\n", i)
	}
	sb.WriteString(")\n")
	return sb.String()
}

func TestRejectsExpandingConstBlocks(t *testing.T) {
	t.Parallel()

	var m *Machine
	// 3,912 bytes converted to 500,501 nodes.
	block := constBlock(500, "= ")
	require.ErrorContains(t, tcCopies("package grid\n"+block), "declarations copy")
	_, err := m.ParseStmts(block)
	require.ErrorContains(t, err, "declarations copy")

	// A 16,920-byte shared type was walked once per spec, 8,008,004 visits.
	block = strings.Replace(constBlock(2000, "["), "+1\n", "+1]int = 1\n", 1)
	_, err = m.ParseExpr("func(){ " + block + "}()")
	require.ErrorContains(t, err, "declarations copy")
}

// Each bare spec gets its own copy of the values, which preprocess rewrites per
// spec (iota). The type node is shared, as before this bound: constCopies counts
// it once per bare spec, so a walk over the shared node is bounded the same way.
func TestConstSpecsOwnTheirCopies(t *testing.T) {
	t.Parallel()

	var m *Machine
	fn, err := m.ParseFile("ok.gno", "package ok\nconst (\n\ta [2]int = [2]int{}\n\tb\n)")
	require.NoError(t, err)
	a := fn.Decls[0].(*ValueDecl)
	b := fn.Decls[1].(*ValueDecl)
	assert.Same(t, a.Type, b.Type)
	assert.NotSame(t, a.Values[0], b.Values[0])
}

// Packages stored before the bound must still load at boot.
func TestBootLoadsLegacyCopies(t *testing.T) {
	t.Parallel()

	body := "package grid\n\ntype Board struct{ a, b, c, d, e, f, g, h struct{ x, y, z struct{ r, g int } } }\n"
	require.ErrorContains(t, tcCopies(body), "declarations copy")

	tm2Store := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	st.AddMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "grid",
		Path:  "gno.land/r/demo/grid",
		Files: []*std.MemFile{{Name: "grid.gno", Body: body}},
	}, MPUserProd)
	m := NewMachineWithOptions(MachineOptions{Store: st, Output: io.Discard})
	require.NotPanics(t, m.PreprocessAllFilesAndSaveBlockNodes)
}

// A const block the type check now refuses must still load at boot, where
// ParseFile re-parses stored packages with no bound of its own.
func TestBootLoadsLegacyConstInheritance(t *testing.T) {
	t.Parallel()

	body := "package grid\n\n" + inheritedConstBlock(50, 99)
	require.ErrorContains(t, tcCopies(body), "declarations copy")

	tm2Store := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	st.AddMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "grid",
		Path:  "gno.land/r/demo/grid",
		Files: []*std.MemFile{{Name: "grid.gno", Body: body}},
	}, MPUserProd)
	m := NewMachineWithOptions(MachineOptions{Store: st, Output: io.Discard})
	require.NotPanics(t, m.PreprocessAllFilesAndSaveBlockNodes)
}

func TestParseAcceptsNameLists(t *testing.T) {
	t.Parallel()

	var m *Machine
	fn, err := m.ParseFile("ok.gno", "package ok\ntype X "+nestedNameLists(2, 4))
	require.NoError(t, err)
	fields := fn.Decls[0].(*TypeDecl).Type.(*StructTypeExpr).Fields
	require.Len(t, fields, 4)
	assert.NotSame(t, fields[0].Type, fields[1].Type, "each name gets its own copy")
}

func TestCheckCopiesBoundary(t *testing.T) {
	t.Parallel()

	// 4 names over 6 levels add 17703 nodes.
	x, err := goparser.ParseExpr(nestedNameLists(6, 4))
	require.NoError(t, err)
	require.NoError(t, checkCopies(x, 17703))
	require.ErrorContains(t, checkCopies(x, 17702), "declarations copy")
}

func TestCopiedNodes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		decl  string
		added int
	}{
		{"type T struct{ a int }", 0},
		{"type T struct{ int }", 1},
		{"type T struct{ a, b int }", 2},
		{"type T struct{ a, b int `t` }", 3},
		{"type T struct{ a, b int // c\n}", 2},
		{"type T func(a, b int) (c, d int)", 4},
		{"type T interface{ M(a, b int) }", 2},
		{"type T struct{ a, b struct{ c, d int } }", 11},
		{"const ( a = 1 + 2; b )", 3},
		{"const ( a int = 1; b )", 2},
		{"const ( a, b = 1, 2; c; d, e )", 0}, // toDecls drops the values
		{"const ( a = struct{ x, y int }{}; b )", 11},
		{"var ( a = 1 + 2; b int )", 0},
	} {
		f, err := goparser.ParseFile(token.NewFileSet(), "", "package p\n"+tc.decl, goparser.ParseComments)
		require.NoError(t, err, tc.decl)
		_, added := copiedNodes(f.Decls[0], 1<<30)
		assert.Equal(t, tc.added, added, tc.decl)
	}
}

// copiedNodes must track Go2Gno. A copy it misses grows the gno AST past the
// go/ast plus the counted copies, on the synthetic sources below or on any
// corpus file that repeats what the new copy duplicates.
func TestCopiedNodesTracksGo2Gno(t *testing.T) {
	t.Parallel()

	sources := map[string]string{
		"names.gno": "package p\ntype X " + nestedNameLists(3, 3),
		"const.gno": "package p\n" + constBlock(30, "= "),
	}
	for _, root := range []string{"../../../examples", "../../stdlibs", "../../tests/files"} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".gno") {
				body, err := os.ReadFile(path)
				sources[path] = string(body)
				return err
			}
			return err
		})
		require.NoError(t, err)
	}

	var m *Machine
	for path, body := range sources {
		f, err := goparser.ParseFile(token.NewFileSet(), path, body, goparser.ParseComments)
		if err != nil {
			continue // negative filetests
		}
		fn, err := m.ParseFile(path, body)
		if err != nil {
			continue
		}
		gnoNodes := 0
		Transcribe(fn, func(_ []Node, _ TransField, _ int, n Node, stage TransStage) (Node, TransCtrl) {
			if stage == TRANS_ENTER {
				gnoNodes++
			}
			return n, TRANS_CONTINUE
		})
		size, added := copiedNodes(f, math.MaxInt/2)
		// Go2Gno adds a few implicit nodes, up to 1.09x in small files.
		assert.LessOrEqual(t, gnoNodes, 2*(size+added), path)
	}
}
