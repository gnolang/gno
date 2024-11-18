package v1

import (
	"std"

	"gno.land/r/demo/upgradable/admin"
)

var store admin.Store = nil

func Init() {
	admin.RegisterLogic(&logic{}, func(_store admin.Store) { store = _store })
}

var _ admin.Logic = &logic{}

type logic struct {}

func (l *logic) ReadCounter() uint64 {
	return store.GetCounter()
}

func (l *logic) UpdateCounter() {
	store.SetCounter(store.GetCounter() + 1)
}

func (l *logic) DebugPrevRealm() string {
	return std.PrevRealm().Addr().String()
}
