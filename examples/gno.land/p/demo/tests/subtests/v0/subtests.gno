package subtests

import (
	"chain/runtime"
	"chain/runtime/unsafe"
)

func GetCurrentRealm() runtime.Realm {
	return unsafe.CurrentRealm()
}

func GetPreviousRealm() runtime.Realm {
	return unsafe.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}
