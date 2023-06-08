package administrable

import (
	"fmt"
	"std"
)


func New() *Administrable {
	return NewWithAddress(std.GetOrigCaller())
}

func NewWithAddress(addr std.Address) *Administrable {
	a := Administrable{
		addrs: make(map[std.Address]struct{}),
	}
	a.add(addr)
	return &a
}

type Administrable struct {
	addrs map[std.Address]struct{}
}

func (a *Administrable) List() []std.Address{
	list := make([]std.Address, len(a.addrs))
	idx := 0
	for addr := range a.addrs {
		list[idx] = addr
		idx++
	}
	return list
}

func (a *Administrable) add(addr std.Address) {
	// XXX: should we panic if addr was already present?
	a.addrs[addr] = struct{}{}
}

func (a *Administrable) del(addr std.Address) {
	// XXX: should we prevent deleting self?
	if len(a.addrs) == 0 {
		panic("cannot have no admin.")
	}
	delete(a.addrs, addr)
}

func (a *Administrable) Has(addr std.Address) bool {
	_, found := a.addrs[addr]
	return found
}

func (a *Administrable) IsAdmin() bool {
	if a.Has(std.GetOrigCaller()){
		return true
	}

	// XXX: also check for std.PrevRealm, when merged.

	return false
}

// AssertIsAdmin checks wether the std.GetOrigCaller or std.PrevRealm is whitelisted.
func (a *Administrable) AssertIsAdmin() {
	if !a.IsAdmin() {
		panic("restricted area.")
	}
}

func (a *Administrable) Add(addr std.Address) {
	a.AssertIsAdmin()
	a.add(addr)
}

func (a *Administrable) Del(addr std.Address) {
	a.AssertIsAdmin()
	a.del(addr)
}
