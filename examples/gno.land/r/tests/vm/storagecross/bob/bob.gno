package bob

import (
	"chain/runtime"
)

func Do(fn func()) {
	println(
		"In bob.Do(), CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
	)
	// Normal non-crossing function doesn't change the current or
	// previous realm.
	// Should behave the same as if /p/bob.Do(), or the same as if
	// fn() directly.
	fn()

}

var private int

func PrivateFunc() {
	// This function is exposed but is not crossing, so fails
	// unless current realm is bob.
	// Generally not useful to expose such a "private function",
	// but can be useful in some circumstances.
	// TODO document when it is useful.
	private = -1
	println(
		"In bob.PrivateFunc(), CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
		"private:", private,
	)
}

type Foo interface {
	Method()
}
type Object struct{}

func (obj *Object) Method() { // CASE rB1, CASE pB1
	println(
		"In bob.Object.Method(), CurrentRealm:", runtime.CurrentRealm().PkgPath(),
		"PreviousRealm:", runtime.PreviousRealm(),
	)
	PrivateFunc()

}

var object Foo

func SetObject(obj Foo) {
	object = obj
} // CASE rC1, CASE pC1
var closure func()

func SetClosure(cls func()) { closure = cls } // CASE rC3, CASE pC3
