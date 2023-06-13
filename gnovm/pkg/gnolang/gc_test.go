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
	vp3 := NewValuePathField(0, 0, "obj3")
	vp2 := NewValuePathField(0, 0, "obj2")
	vp1 := NewValuePathField(0, 0, "obj1")

	obj3 := &GCObj{paths: []*ValuePath{&vp3}}
	obj2 := &GCObj{paths: []*ValuePath{&vp2}, ref: obj3}
	obj1 := &GCObj{paths: []*ValuePath{&vp1}, ref: obj2}

	// Create garbage collector
	gc := NewGC(true)

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	gc.AddRoot(obj1)

	// Collect garbage
	gc.Collect()
	assert.NotNil(t, gc.getObjByPath(obj1.paths[0]))
	assert.NotNil(t, gc.getObjByPath(obj2.paths[0]))
	assert.NotNil(t, gc.getObjByPath(obj3.paths[0]))
}

func TestGC_RemoveRoot(t *testing.T) {
	vp3 := NewValuePathField(0, 0, "obj3")
	vp2 := NewValuePathField(0, 0, "obj2")
	vp1 := NewValuePathField(0, 0, "obj1")

	obj3 := &GCObj{paths: []*ValuePath{&vp3}}
	obj2 := &GCObj{paths: []*ValuePath{&vp2}, ref: obj3}
	obj1 := &GCObj{paths: []*ValuePath{&vp1}, ref: obj2}

	// Create garbage collector
	gc := NewGC(true)

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	gc.AddRoot(obj1)

	// Collect garbage
	gc.Collect()
	assert.NotNil(t, gc.getObjByPath(obj1.paths[0]))
	assert.NotNil(t, gc.getObjByPath(obj2.paths[0]))
	assert.NotNil(t, gc.getObjByPath(obj3.paths[0]))

	gc.RemoveRoot(obj1.paths[0])
	gc.Collect()

	assert.Nil(t, gc.getObjByPath(obj1.paths[0]))
	assert.Nil(t, gc.getObjByPath(obj2.paths[0]))
	assert.Nil(t, gc.getObjByPath(obj3.paths[0]))
}

func TestGC_CollectUnsedObjects(t *testing.T) {
	vp3 := NewValuePathField(0, 0, "obj3")
	vp2 := NewValuePathField(0, 0, "obj2")
	vp1 := NewValuePathField(0, 0, "obj1")

	obj3 := &GCObj{paths: []*ValuePath{&vp3}}
	obj2 := &GCObj{paths: []*ValuePath{&vp2}, ref: obj3}
	obj1 := &GCObj{paths: []*ValuePath{&vp1}, ref: obj2}

	// Create garbage collector
	gc := NewGC(true)

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	// Collect garbage
	gc.Collect()
	assert.Nil(t, gc.getObjByPath(obj1.paths[0]))
	assert.Nil(t, gc.getObjByPath(obj2.paths[0]))
	assert.Nil(t, gc.getObjByPath(obj3.paths[0]))
	assert.Empty(t, gc.objs)
	assert.Empty(t, gc.roots)
}
