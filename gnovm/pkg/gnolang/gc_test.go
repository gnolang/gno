package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkEscapeAnalysis(b *testing.B) {
	f, err := parser.ParseFile(token.NewFileSet(), "",
		`
		package p
		func foo() {
			a := 5
			b := &a // both should escape
			c := b // should escape
			e := 5 // should not escape
			aa := c
			bb := aa
			cc := bb
			dd := cc
			ee := dd
			ff := ee
			gg := ff
			hh := gg
			ii := hh
			jj := ii
			kk := jj
			ll := kk
			mm := ll
			nn := mm
			oo := nn
			pp := oo
			qq := pp
			rr := qq
			ss := rr
			tt := ss

		}`, 0)
	require.NoError(b, err)

	fn := f.Decls[0].(*ast.FuncDecl)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		escapedVars := EscapeAnalysis(fn)

		b.StopTimer()
		require.Len(b, escapedVars, 23)
		b.StartTimer()
	}

}

func TestGC_NotCollectUsedObjects(t *testing.T) {
	obj3 := &GCObj{path: "obj3"}
	obj2 := &GCObj{path: "obj2", ref: obj3}
	obj1 := &GCObj{path: "obj1", ref: obj2}

	// Create garbage collector
	gc := NewGC()

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	gc.AddRoot(obj1)

	// Collect garbage
	gc.Collect()
	assert.NotNil(t, gc.getObjByPath(obj1.path))
	assert.NotNil(t, gc.getObjByPath(obj2.path))
	assert.NotNil(t, gc.getObjByPath(obj3.path))
}

func TestGC_RemoveRoot(t *testing.T) {
	obj3 := &GCObj{path: "obj3"}
	obj2 := &GCObj{path: "obj2", ref: obj3}
	obj1 := &GCObj{path: "obj1", ref: obj2}

	// Create garbage collector
	gc := NewGC()

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	gc.AddRoot(obj1)

	// Collect garbage
	gc.Collect()
	assert.NotNil(t, gc.getObjByPath(obj1.path))
	assert.NotNil(t, gc.getObjByPath(obj2.path))
	assert.NotNil(t, gc.getObjByPath(obj3.path))

	gc.RemoveRoot(obj1.path)
	gc.Collect()

	assert.Nil(t, gc.getObjByPath(obj1.path))
	assert.Nil(t, gc.getObjByPath(obj2.path))
	assert.Nil(t, gc.getObjByPath(obj3.path))
}

func TestGC_CollectUnsedObjects(t *testing.T) {
	obj3 := &GCObj{path: "obj3"}
	obj2 := &GCObj{path: "obj2", ref: obj3}
	obj1 := &GCObj{path: "obj1", ref: obj2}

	// Create garbage collector
	gc := NewGC()

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	// Collect garbage
	gc.Collect()
	assert.Nil(t, gc.getObjByPath(obj1.path))
	assert.Nil(t, gc.getObjByPath(obj2.path))
	assert.Nil(t, gc.getObjByPath(obj3.path))
	assert.Empty(t, gc.objs)
	assert.Empty(t, gc.roots)
}

type escapeTest struct {
	testName     string
	code         string
	declaration  func(*ast.File) *ast.FuncDecl
	expectedVars []string
}

var escapeTests = []escapeTest{
	{
		testName: "variables simple",
		code: `
		package p
		func foo() {
			a := 5
			b := &a // both should escape
			c := b // should escape
			e := 5 // should not escape
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[0].(*ast.FuncDecl)
		},
		expectedVars: []string{"a", "b", "c"},
	},
	{
		testName: "closures",
		code: `
		package p
		func foo() {
			a := 5
			func(c int)  int {
				b := a // both should escape
				return c
			}
			e := 5 // should not escape
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[0].(*ast.FuncDecl)
		},
		expectedVars: []string{"a", "b"},
	},
	{
		testName: "special built-in types",
		code: `
		package p
		func foo(x string) {
			func(a, d string, b map[string]bool, c []string, e int) {
			}
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[0].(*ast.FuncDecl)
		},
		expectedVars: []string{"a", "b", "c", "d", "x"},
	},
	{
		testName: "goroutines",
		code: `
		package main
		
		func f() {
			i := 1
			go func() {
				_ = i
			}()
			b := 4
			foo(&b)
		}
		`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			var out *ast.FuncDecl
			for _, decl := range f.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "f" {
					out = fn
					break
				}
			}

			return out
		},
		expectedVars: []string{"i", "b"},
	},
}

func TestEscapeAnalysis(t *testing.T) {
	for _, et := range escapeTests {
		t.Run(et.testName, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "", et.code, 0)
			require.NoError(t, err)
			fn := et.declaration(f)
			ev := EscapeAnalysis(fn)
			assert.ElementsMatch(t, ev, et.expectedVars)
		})
	}
}
