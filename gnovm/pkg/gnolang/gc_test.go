package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGC_NotCollectUsedObjects(t *testing.T) {
	obj1 := &GCObj{id: ObjectIDFromPkgPath("obj1")}
	obj2 := &GCObj{id: ObjectIDFromPkgPath("obj2")}
	obj3 := &GCObj{id: ObjectIDFromPkgPath("obj3")}

	// Link objects together
	obj1.AddRef(obj2)
	obj2.AddRef(obj3)

	// Create garbage collector
	gc := NewGC()

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	gc.AddRoot(obj1)

	// Collect garbage
	gc.Collect()
	assert.NotNil(t, gc.getObj(obj1.id))
	assert.NotNil(t, gc.getObj(obj2.id))
	assert.NotNil(t, gc.getObj(obj3.id))
}

func TestGC_CollectUnsedObjects(t *testing.T) {
	obj1 := &GCObj{id: ObjectIDFromPkgPath("obj1")}
	obj2 := &GCObj{id: ObjectIDFromPkgPath("obj2")}
	obj3 := &GCObj{id: ObjectIDFromPkgPath("obj3")}

	// Link objects together
	obj1.AddRef(obj2)
	obj2.AddRef(obj3)

	// Create garbage collector
	gc := NewGC()

	// Add objects to garbage collector
	gc.AddObject(obj1)
	gc.AddObject(obj2)
	gc.AddObject(obj3)

	// Collect garbage
	gc.Collect()
	assert.Nil(t, gc.getObj(obj1.id))
	assert.Nil(t, gc.getObj(obj2.id))
	assert.Nil(t, gc.getObj(obj3.id))
	assert.Empty(t, gc.objs)
	assert.Empty(t, gc.roots)
}
