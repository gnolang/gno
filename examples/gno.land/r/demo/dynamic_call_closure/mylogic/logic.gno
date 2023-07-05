package mylogic

import (
	"std"

	"gno.land/r/demo/dynamic_call_closure/myproxy"
	"gno.land/r/demo/tests" // TODO: replace tests demo with foo20 for more realism
)

// unsafe execute for demo
func ExecuteApprovedProposals() {
	assertConditionsAreMet()
	println("AAAA1 calling tests.GetPrevRealm", tests.GetPrevRealm().PkgPath())
	println("AAAA2 calling std.PrevRealm", std.PrevRealm().PkgPath())

	// myproxy.ExecAsMe()(myproxy.MyLogicInMyProxyClosure(42)) // works, but i need to have all my logic in closure when publishing the closure
	myproxy.MyLogicInMyProxy(42) // useless proxy here

	myproxy.ExecAsMe()(func() {
		println("BBBB1 calling tests.GetPrevRealm", tests.GetPrevRealm().PkgPath()) // returning logic instead of proxy
		println("BBBB2 calling std.PrevRealm", std.PrevRealm().PkgPath())
	})

	myFoo20 := foo20.GetFoo20AsCaller()
	myFoo20.Transfer(to, 4242) // from is mylogic  (current realm?)

}

// expected (by manfred) vs actual behavior
// mylogic > myproxy (closure stays in the same realm) > foo20                     // what I expect
// mylogic > myproxy                                   > mylogic (closure) > foo20 // what we have

func assertConditionsAreMet() {
	// here you should restrict who can call the contract.
}
