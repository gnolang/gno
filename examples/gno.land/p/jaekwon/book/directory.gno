package book

import (
	"gno.land/r/demo/avl"
)

//----------------------------------------
// Directory

type Directory struct {
	name    string   // name of directory
	objects avl.Tree // key -> Object
	auto    bool     // use automatic numeric keys
	lastKey int64    // last auto numeric key used
	prefix  string   // key prefix, if any
}

func NewDirectory(name string, auto bool, prefix string) *Directory {
	return &Directory{
		name:    name,
		objects: nil,
		auto:    auto,
		lastKey: 0,
		prefix:  prefix,
	}
}

func (dir *Directory) GetName() string {
	return dir.name
}

func (dir *Directory) GetObject(key ObjectKey) Object {
	obj, exists := dir.objects.Get(key)
	if !exists {
		panic("object not found: invalid key")
	}
	return obj
}

func (dir *Directory) AddObject(obj Object) {
	key := ObjectKey("")
	if dir.auto {
		key = dir.nextKey()
		obj.SetKey(key)
	} else {
		key = obj.GetKey()
		key.ValidateBasic()
		if !hasPrefix(string(key), dir.prefix) {
			panic("invalid key: prefix mismatch")
		}
	}
	updated := dir.objects.Set(string(key), obj)
	if updated == true {
		panic("duplicate object key")
	}
}

func (dir *Directory) RemoveObject(obj Object) {
	key := obj.GetKey()
	old, removed := dir.objects.Remove(string(key))
	if !removed {
		panic("cannot remove object: not found")
	}
	if old != obj {
		panic("cannot remove conflicting object")
	}
}

func (dir *Directory) RemoveObjectByKey(key ObjectKey) {
	if !hasPrefix(key, dir.prefix) {
		panic("invalid key: prefix mismatch")
	}
	_, removed := dir.objects.Remove(string(key))
	if !removed {
		panic("cannot remove object: not found")
	}
}

func (dir *Directory) nextKey() ObjectKey {
	if !dir.auto {
		panic("cannot generate key: not auto keyed")
	}
	nextKey := dir.lastKey + 1
	if nextKey < 0 {
		panic("key overflow")
	}
	dir.lastKey = nextKey
	return ObjectKey(dir.prefix + itoa(nextKey))
}
