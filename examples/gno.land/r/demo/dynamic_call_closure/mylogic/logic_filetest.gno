// PKGPATH: gno.land/r/demo/mylogic_test
package mylogic_test

import "gno.land/r/demo/dynamic_call_closure/mylogic"
import "std"

func main() {
	mylogic.ExecuteApprovedProposals()
	println("DONE")
}

// Output:
// AAAA1 calling tests.GetPrevRealm gno.land/r/demo/dynamic_call_closure/mylogic
// AAAA2 calling std.PrevRealm gno.land/r/demo/mylogic_test
// DDDD1 calling tests.GetPrevRealm gno.land/r/demo/dynamic_call_closure/myproxy 42
// DDDD2 calling std.PrevRealm gno.land/r/demo/dynamic_call_closure/mylogic 42
// BBBB1 calling tests.GetPrevRealm gno.land/r/demo/dynamic_call_closure/mylogic
// BBBB2 calling std.PrevRealm gno.land/r/demo/dynamic_call_closure/myproxy
// DONE
