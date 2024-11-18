package admin

import (
	"std"
)

type Logic interface {
	ReadCounter() uint64
	UpdateCounter()

	DebugPrevRealm() string
}

func RegisterLogic(l Logic, setStore func(Store)) {
	ReadCounter = l.ReadCounter
	UpdateCounter = l.UpdateCounter
	DebugPrevRealm = l.DebugPrevRealm

	newStore := store()
	setStore(newStore)
}