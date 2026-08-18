package subtests

import (
	"chain/runtime"
	"chain/runtime/unsafe"
)

func GetCurrentRealm(cur realm) runtime.Realm {
	return unsafe.CurrentRealm()
}

func GetPreviousRealm(cur realm) runtime.Realm {
	return unsafe.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func CallAssertOriginCall(cur realm) {
	runtime.AssertOriginCall()
}

func CallIsOriginCall(cur realm) bool {
	return unsafe.PreviousRealm().IsUser()
}

func BankerOriginSend(cur realm) string {
	return unsafe.OriginSend().String()
}
