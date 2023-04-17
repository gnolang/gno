package gnolang

import (
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
