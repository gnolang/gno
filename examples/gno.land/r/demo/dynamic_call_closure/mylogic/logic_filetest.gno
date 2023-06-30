// PKGPATH: gno.land/r/demo/mylogic_test
package mylogic_test

import "gno.land/r/demo/dynamic_call_closure/mylogic"
import "std"

func main() {
	mylogic.ProposeTransfer(std.Address("foo"), 1337000)
	mylogic.ExecuteApprovedProposals()
	mylogic.ApproveTransfers()
	mylogic.ApproveTransfers()
	mylogic.ApproveTransfers()
	mylogic.ApproveTransfers()
	mylogic.ApproveTransfers()
	mylogic.ExecuteApprovedProposals()
	println("DONE")
}


// Output:
// calling tests.GetPrevRealm gno.land/r/demo/dynamic_call_closure/mylogic
// proposal parameters are -> to: foo , amount: 1337000
// DONE
