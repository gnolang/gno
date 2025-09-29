package subtests

import "chain/runtime"

func GetCurrentRealm(cur realm) runtime.Realm {
	return runtime.CurrentRealm()
}

func GetPreviousRealm(cur realm) runtime.Realm {
	return runtime.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func CallAssertOriginCall(cur realm) {
	runtime.AssertOriginCall()
}

func CallIsOriginCall(cur realm) bool {
	return runtime.PreviousRealm().IsUser()
}
