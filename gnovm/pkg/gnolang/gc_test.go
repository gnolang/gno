package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGC_NotCollectUsedObjects(t *testing.T) {
	obj1 := &GCObj{key: MapKey("obj1")}
	obj2 := &GCObj{key: MapKey("obj2")}
	obj3 := &GCObj{key: MapKey("obj3")}

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
	assert.NotNil(t, gc.getObj(obj1.key))
	assert.NotNil(t, gc.getObj(obj2.key))
	assert.NotNil(t, gc.getObj(obj3.key))
}

func TestGC_CollectUnsedObjects(t *testing.T) {
	obj1 := &GCObj{key: MapKey("obj1")}
	obj2 := &GCObj{key: MapKey("obj2")}
	obj3 := &GCObj{key: MapKey("obj3")}

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
	assert.Nil(t, gc.getObj(obj1.key))
	assert.Nil(t, gc.getObj(obj2.key))
	assert.Nil(t, gc.getObj(obj3.key))
	assert.Empty(t, gc.objs)
	assert.Empty(t, gc.roots)
}
