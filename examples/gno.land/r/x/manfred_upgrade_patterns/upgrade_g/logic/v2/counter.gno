package v2

import (
	"std"

	"gno.land/r/x/manfred_upgrade_patterns/upgrade_g/admin"
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
	store.SetCounter(store.GetCounter() + 2)
}

func (l *logic) DebugPrevRealm() string {
	return std.PrevRealm().Addr().String()
}
