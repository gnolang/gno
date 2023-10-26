package book

import (
	"std"

	"gno.land/r/demo/avl"
)

//----------------------------------------
// Directory

type Directory struct {
	name    string   // name of directory
	objects avl.Tree // std.Address -> Addressable
}

func NewDirectory(name string) *Directory {
	return &Directory{
		name:    name,
		objects: nil,
	}
}

func (dir *Directory) GetName() string {
	return dir.name
}

func (dir *Directory) Get(addr std.Address) std.Addressable {
	obj, exists := dir.objects.Get(addr)
	if !exists {
		panic("object not found: invalid address")
	}
	return obj
}

func (dir *Directory) Has(addr std.Address) bool {
	return dir.objects.Has(addr)
}

func (dir *Directory) Add(obj std.Addressable) {
	addr := obj.GetAddress()
	updated := dir.objects.Set(string(addr), obj)
	if updated == true {
		panic("duplicate address")
	}
}

func (dir *Directory) Remove(obj std.Addressable) {
	addr := obj.GetAddress()
	old, removed := dir.objects.Remove(string(addr))
	if !removed {
		panic("cannot remove object: not found")
	}
	if old != obj {
		panic("cannot remove conflicting object")
	}
}

func (dir *Directory) RemoveByAddress(addr std.Address) {
	_, removed := dir.objects.Remove(string(addr))
	if !removed {
		panic("cannot remove object: not found")
	}
}
