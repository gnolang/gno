package store

import (
	"std"
	"fmt"

	"gno.land/r/demo/upgradable/admin"
)

var counterValue uint64

var _ admin.Store = &store{}

var currentStore *store

type store struct {}

func (s *store) GetCounter() uint64 {
	if s != currentStore {
		panic("Revoked store")
	}

	return counterValue
}

func (s *store) SetCounter(value uint64) {
	if s != currentStore {
		panic("Revoked store")
	}

	counterValue = value
}

func newStore() admin.Store {
	currentStore = &store{}

	return currentStore
}

func Init() {
	admin.RegisterStore(newStore)
}