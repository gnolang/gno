package myproxy

import (
	"gno.land/r/demo/tests"
	"std"
)

// no import needed.

func MyLogicInMyProxyClosure(amount int) func() {
	return func() {
		println("CCCC1 calling tests.GetPrevRealm", tests.GetPrevRealm().PkgPath(), amount)
		println("CCCC2 calling std.PrevRealm", std.PrevRealm().PkgPath(), amount)
	}
}

func MyLogicInMyProxy(amount int) {
	println("DDDD1 calling tests.GetPrevRealm", tests.GetPrevRealm().PkgPath(), amount)
	println("DDDD2 calling std.PrevRealm", std.PrevRealm().PkgPath(), amount)
}

func ExecAsMeV1(cb func()) {
	assertConditionsAreMet()
	cb()
}

func ExecAsMeV2() func(func()) {
	assertConditionsAreMet()
	foo20.Transfer()
	return func(cb func()) {
		cb()
	}
	//MyLogicInMyProxy(42)
	//MyLogicInMyProxyClosure()(42)
}

func assertConditionsAreMet() {
	// here we can make assertions for security.
	// std.GetOrigCaller()
	// std.PrevRealm()
	// std.GetCallerAt(2)
	// ...
}
