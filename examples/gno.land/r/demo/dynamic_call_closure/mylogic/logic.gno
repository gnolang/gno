package mylogic

import (
	"std"

	"gno.land/r/demo/dynamic_call_closure/myproxy"
	"gno.land/r/demo/tests" // TODO: replace tests demo with foo20 for more realism
)

type proposal struct{
	to std.Address
	amount int
	approvals int
}

var proposals []*proposal

// unsafe proposal for demo
func ProposeTransfer(to std.Address, amount int) {
	assertConditionsAreMet()

	proposals = append(proposals, &proposal{to: to, amount: amount, approvals: 1})
}

// unsafe approval for demo
func ApproveTransfers() {
	assertConditionsAreMet()

	for _, proposal := range proposals {
		proposal.approvals++
	}
}

// unsafe execute for demo
func ExecuteApprovedProposals() {
	assertConditionsAreMet()

	for _, proposal := range proposals {
		if proposal.approvals < 3 {
			continue
		}
		myproxy.ExecAsMe(func() {
			//foo20.Transfer(proposal.to, proposal.amount)
			println("calling tests.GetPrevRealm", tests.GetPrevRealm().PkgPath())
			println("proposal parameters are -> to:", proposal.to, ", amount:", proposal.amount)
		})
		// TODO: GC passed approvals
	}
}

func assertConditionsAreMet() {
	// here you should restrict who can call the contract.
}
