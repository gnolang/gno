package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				b := &a
				c := b
			}`, 0)
	if err != nil {
		t.Fatal(err)
	}

	fn := f.Decls[0].(*ast.FuncDecl)
	escapedVars := EscapeAnalysis(fn)

	assert.ElementsMatch(t, escapedVars, []string{"a", "b", "c"})
}
