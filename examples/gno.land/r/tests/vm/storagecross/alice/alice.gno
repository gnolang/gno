package alice

import (
	"chain/runtime"

	"gno.land/r/tests/vm/storagecross/bob"
)

func TopFunc() { // CASE rA1
	println(
		"In alice.TopFunc(),",
		"CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
	)

	bob.PrivateFunc()

}
func TopFuncFnNoncrossing() func() {
	println(
		"In alice.TopFuncFnNoncrossing(),",
		"CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
	)

	return func() { // CASE rA2
		println(
			"In alice.TopFuncFnNoncrossing().inner,",
			"CurrentRealm:", runtime.CurrentRealm().PkgPath(),
			"PreviousRealm:", runtime.PreviousRealm(),
		)
		bob.PrivateFunc()
	}
}
func TopFuncFnCrossing(cur realm) func() {
	println(
		"In alice.TopFuncFnCrossing(),",
		"CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
	)
	return func() { // CASE rA3
		println(
			"In alice.TopFuncFnCrossing().inner,",
			"CurrentRealm:", runtime.CurrentRealm().PkgPath(),
			"PreviousRealm:", runtime.PreviousRealm(),
		)
		bob.PrivateFunc()
	}
}

type Object struct{}

func (obj *Object) Method() { // CASE rB2~4, CASE rC1~5

	println(
		"In alice.Object.Method(), CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
	)
	bob.PrivateFunc()
}

var object *Object

func SetObject(obj *Object) { object = obj } // CASE rC2, CASE pC2

var closure func()

func SetClosure(cls func()) { closure = cls } // case rC4~5, CASE pC4~5
