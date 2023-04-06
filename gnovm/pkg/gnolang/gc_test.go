package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGC_NotCollectUsedObjects(t *testing.T) {
	obj1 := &GCObj{data: defaultStructValue(nil, &StructType{
		PkgPath: "",
		Fields:  nil,
		typeid:  "",
	})}
	obj2 := &GCObj{data: defaultStructValue(nil, &StructType{
		PkgPath: "",
		Fields:  nil,
		typeid:  "",
	})}
	obj3 := &GCObj{data: defaultStructValue(nil, &StructType{
		PkgPath: "",
		Fields:  nil,
		typeid:  "",
	})}

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
	assert.NotNil(t, gc.getObj(obj1.data.GetObjectID()))
	assert.NotNil(t, gc.getObj(obj1.data.GetObjectID()))
	assert.NotNil(t, gc.getObj(obj1.data.GetObjectID()))
}

func TestGC_CollectUnsedObjects(t *testing.T) {
	obj1 := &GCObj{data: defaultStructValue(nil, &StructType{
		PkgPath: "",
		Fields:  nil,
		typeid:  "",
	})}
	obj2 := &GCObj{data: defaultStructValue(nil, &StructType{
		PkgPath: "",
		Fields:  nil,
		typeid:  "",
	})}
	obj3 := &GCObj{data: defaultStructValue(nil, &StructType{
		PkgPath: "",
		Fields:  nil,
		typeid:  "",
	})}

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
	assert.Nil(t, gc.getObj(obj1.data.GetObjectID()))
	assert.Nil(t, gc.getObj(obj1.data.GetObjectID()))
	assert.Nil(t, gc.getObj(obj1.data.GetObjectID()))
	assert.Empty(t, gc.objs)
	assert.Empty(t, gc.roots)
}
