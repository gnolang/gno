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

func TestEscapeAnalysis(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "",
		`
			package p
			func foo() {
				a := 5
				b := &a // both should escape
				c := b // should escape
				e := 5 // should not escape
			}`, 0)
	require.NoError(t, err)

	fn := f.Decls[0].(*ast.FuncDecl)
	escapedVars := EscapeAnalysis(fn)

	assert.ElementsMatch(t, escapedVars, []string{"a", "b", "c"})
}

func TestEscapeAnalysisClosure(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "",
		`
			package p
			func foo() {
				a := 5
				
				func(c int)  int {
					b := a // both should escape

					return c
				}

				e := 5 // should not escape
			}`, 0)
	require.NoError(t, err)

	fn := f.Decls[0].(*ast.FuncDecl)
	escapedVars := EscapeAnalysis(fn)

	assert.ElementsMatch(t, escapedVars, []string{"a", "b"})
}

func TestEscapeAnalysisWithGoroutine(t *testing.T) {
	src := `
package main

func f() {
    i := 1
    go func() {
        _ = i
    }()
	b := 4
	foo(&b)
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err, "Failed to parse source code")

	// Find the function declaration
	var f *ast.FuncDecl
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "f" {
			f = fn
			break
		}
	}
	require.NotNil(t, f, "failed to find function declaration")

	// Test the function
	heapVars := EscapeAnalysis(f)
	expected := []string{"i", "b"}

	require.Equal(t, expected, heapVars)
}
