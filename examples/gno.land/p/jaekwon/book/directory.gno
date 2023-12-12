package book

import (
	"std"

	"gno.land/r/demo/avl"
)

//----------------------------------------
// Diretory

// The Directory as interface is arguably a bit too powerful.
// Custom implementations can do anything, so be careful
// to use only fully vetted, audited, complete implementations.
type Directory interface {
	GetName() string
	Get(std.Address) std.Addressable
	Has(std.Address) bool
	Add(std.Addressable)
	Remove(std.Addressable)
	RemoveByAddress(std.Addressable)
}

//----------------------------------------
// defaultDirectory

type defaultDirectory struct {
	name    string   // name of directory
	objects avl.Tree // std.Address -> Addressable
}

func NewDefaultDirectory(name string) *defaultDirectory {
	return &defaultDirectory{
		name:    name,
		objects: nil,
	}
}

func (dir *defaultDirectory) GetName() string {
	return dir.name
}

func (dir *defaultDirectory) Get(addr std.Address) std.Addressable {
	obj, exists := dir.objects.Get(addr)
	if !exists {
		panic("object not found: invalid address")
	}
	return obj
}

func (dir *defaultDirectory) Has(addr std.Address) bool {
	return dir.objects.Has(addr)
}

func (dir *defaultDirectory) Add(obj std.Addressable) {
	addr := obj.GetAddress()
	updated := dir.objects.Set(string(addr), obj)
	if updated == true {
		panic("duplicate address")
	}
}

func (dir *defaultDirectory) Remove(obj std.Addressable) {
	addr := obj.GetAddress()
	old, removed := dir.objects.Remove(string(addr))
	if !removed {
		panic("cannot remove object: not found")
	}
	if old != obj {
		panic("cannot remove conflicting object")
	}
}

func (dir *defaultDirectory) RemoveByAddress(addr std.Address) {
	_, removed := dir.objects.Remove(string(addr))
	if !removed {
		panic("cannot remove object: not found")
	}
}
