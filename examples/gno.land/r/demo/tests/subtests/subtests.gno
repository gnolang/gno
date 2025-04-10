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

func CallAssertOriginCall() {
	runtime.AssertOriginCall()
}

func CallIsOriginCall() bool {
	return runtime.PreviousRealm().IsUser()
}
