package gnolang

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// TestPredefineDeclIndexMatchesFileSet checks that both lookup paths select the same declarations.
func TestPredefineDeclIndexMatchesFileSet(t *testing.T) {
	sharedFirst := &ValueDecl{NameExprs: NameExprs{{Name: "shared"}}}
	sharedSecond := &TypeDecl{NameExpr: NameExpr{Name: "shared"}}
	multi := &ValueDecl{NameExprs: NameExprs{{Name: "a"}, {Name: "b"}}}
	imported := &ValueDecl{NameExprs: NameExprs{{Name: "imported"}}}
	method := &FuncDecl{NameExpr: NameExpr{Name: "method"}, IsMethod: true}
	fset := &FileSet{Files: []*FileNode{
		{
			FileName: "first.gno",
			Decls: Decls{
				&ValueDecl{NameExprs: NameExprs{{Name: "shared"}}},
				&ImportDecl{NameExpr: NameExpr{Name: "imported"}},
				imported,
				&ImportDecl{NameExpr: NameExpr{Name: "importOnly"}},
				multi,
				&ValueDecl{NameExprs: NameExprs{{Name: blankIdentifier}}},
				method,
			},
		},
		{
			FileName: "last.gno",
			Decls:    Decls{sharedFirst, sharedSecond},
		},
	}}

	index := newPredefineDeclIndex(fset)
	tests := []struct {
		name Name
		decl Decl
	}{
		{"shared", sharedFirst}, // last file, first declaration in that file
		{"a", multi},
		{"b", multi},
		{"imported", imported},
		{"importOnly", nil},
		{blankIdentifier, nil},
		{"method", nil},
		{"missing", nil},
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			wantFile, wantSlot, wantOK := fset.GetDeclForSafe(tt.name)
			if wantOK != (tt.decl != nil) {
				t.Fatalf("invalid fixture: GetDeclForSafe(%q) ok = %v", tt.name, wantOK)
			}
			if wantOK && *wantSlot != tt.decl {
				t.Fatalf("invalid fixture: GetDeclForSafe(%q) returned %T %p, want %T %p", tt.name, *wantSlot, *wantSlot, tt.decl, tt.decl)
			}

			got, gotOK := index.entries[tt.name]
			if gotOK != wantOK {
				t.Fatalf("index[%q] ok = %v, want %v", tt.name, gotOK, wantOK)
			}
			if gotOK && (got.file != wantFile || got.decl != *wantSlot) {
				t.Fatalf("index[%q] = (%p, %T %p), want (%p, %T %p)", tt.name, got.file, got.decl, got.decl, wantFile, *wantSlot, *wantSlot)
			}
		})
	}
}

// TestPredefineDeclIndexPreservesOverrideWhenSplitting checks that a shadowed split stays shadowed.
func TestPredefineDeclIndexPreservesOverrideWhenSplitting(t *testing.T) {
	m := NewMachine("", nil)
	defer m.Release()

	// Replacing the winning a with the shadowed split part would report a
	// false x -> a -> x cycle.
	first := m.MustParseFile("first.gno", `package override
var x, a = a, x
`)
	last := m.MustParseFile("last.gno", `package override
var a = 1
`)
	fset := &FileSet{Files: []*FileNode{first, last}}

	m.PreprocessFiles("override", "test/override", fset, false, true)
}

// TestPredefineReverseChainAllocsAreLinear guards against quadratic allocation growth.
func TestPredefineReverseChainAllocsAreLinear(t *testing.T) {
	if debugAssert {
		t.Skip("debug assertions intentionally perform live declaration scans")
	}

	allocsFor := func(n int) uint64 {
		source := predefineTypeChainSource(n)
		m := NewMachine("", nil)
		defer m.Release()
		file, err := m.ParseFile("chain.gno", source)
		if err != nil {
			t.Fatal(err)
		}
		files := &FileSet{Files: []*FileNode{file}}
		pkg := NewPackageNode("chain", "test/chain", files)

		// Measure only allocations from PredefineFileSet.
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		PredefineFileSet(m.Store, pkg, files)
		runtime.ReadMemStats(&after)
		return after.Mallocs - before.Mallocs
	}

	const n = 1000
	small := allocsFor(n)
	large := allocsFor(2 * n)

	// Quadratic growth approaches 4x when n doubles.
	ratio := float64(large) / float64(small)
	t.Logf("PredefineFileSet mallocs: N=%d -> %d, N=%d -> %d (ratio %.2fx)", n, small, 2*n, large, ratio)
	if ratio > 3 {
		t.Fatalf("PredefineFileSet allocations scaled %.2fx when doubling N (want < 3x); the O(N^2) declaration lookup likely regressed", ratio)
	}
}

func predefineTypeChainSource(n int) string {
	var source strings.Builder
	source.WriteString("package chain\n")
	for i := range n {
		if i < n-1 {
			fmt.Fprintf(&source, "type T%d T%d\n", i, i+1)
		} else {
			fmt.Fprintf(&source, "type T%d int\n", i)
		}
	}
	return source.String()
}

// BenchmarkPredefine measures dependency chains in both declaration orders.
func BenchmarkPredefine(b *testing.B) {
	for _, direction := range []string{"Reverse", "Forward"} {
		for _, n := range []int{250, 500, 1000, 2000} {
			b.Run(fmt.Sprintf("%s/N=%d", direction, n), func(b *testing.B) {
				source := predefineChainSource(n, direction == "Reverse")
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					m := NewMachine("", nil)
					file, err := m.ParseFile("chain.gno", source)
					if err != nil {
						b.Fatal(err)
					}
					files := &FileSet{Files: []*FileNode{file}}
					pkg := NewPackageNode("chain", "benchmark/chain", files)
					PredefineFileSet(m.Store, pkg, files)
					m.Release()
				}
			})
		}
	}
}

// predefineChainSource builds a package with a linear declaration dependency chain.
func predefineChainSource(n int, reverse bool) string {
	var source strings.Builder
	source.WriteString("package chain\n")
	for i := range n {
		if reverse && i < n-1 {
			fmt.Fprintf(&source, "var v%d = v%d\n", i, i+1)
		} else if !reverse && i > 0 {
			fmt.Fprintf(&source, "var v%d = v%d\n", i, i-1)
		} else {
			fmt.Fprintf(&source, "var v%d = 1\n", i)
		}
	}
	return source.String()
}
