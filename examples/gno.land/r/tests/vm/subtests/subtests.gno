package subtests

import (
	"chain/banker"
	"chain/runtime"
)

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

func RealmCoins(cur realm) string {
	return cur.Coins().String()
}

func BankerOriginSend(cur realm) string {
	return banker.OriginSend().String()
}
