package subtests

import "chain/runtime"

func GetCurrentRealm() runtime.Realm {
	return runtime.CurrentRealm()
}

func GetPreviousRealm() runtime.Realm {
	return runtime.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}
