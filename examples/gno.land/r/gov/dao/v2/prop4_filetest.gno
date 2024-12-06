package main

import (
	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
	govdaov2 "gno.land/r/gov/dao/v2"
	"gno.land/r/sys/params"
)

func init() {
	mExec := params.NewStringPropExecutor("prop1.string", "value1")
	title := "Setting prop1.string param"
	comment := "setting prop1.string param"
	prop := dao.ProposalRequest{
		Title:       title,
		Description: comment,
		Executor:    mExec,
	}
	id := bridge.GovDAO().Propose(prop)
	println("new prop", id)
}

func main() {
	println("--")
	println(govdaov2.Render(""))
	println("--")
	println(govdaov2.Render("0"))
	println("--")
	bridge.GovDAO().VoteOnProposal(0, "YES")
	println("--")
	println(govdaov2.Render("0"))
	println("--")
	bridge.GovDAO().ExecuteProposal(0)
	println("--")
	println(govdaov2.Render("0"))
}

// Output:
// new prop 0
// --
// # GovDAO Proposals
//
// ## [Prop #0 - Setting prop1.string param](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// # Proposal #0 - Setting prop1.string param
//
// ## Description
//
// setting prop1.string param
//
// ## Proposal information
//
// **Status: ACTIVE**
//
// **Voting stats:**
// - YES 0 (0%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTES 10 (100%)
//
//
// **Threshold met: FALSE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
// ### Actions
//
// #### [[Vote YES](gno.land/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=YES)] - [[Vote NO](gno.land/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=NO)] - [[Vote ABSTAIN](gno.land/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=ABSTAIN)]
//
//
// --
// --
// # Proposal #0 - Setting prop1.string param
//
// ## Description
//
// setting prop1.string param
//
// ## Proposal information
//
// **Status: ACCEPTED**
//
// **Voting stats:**
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTES 0 (0%)
//
//
// **Threshold met: TRUE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
// ### Actions
//
// The voting period for this proposal is over.
//
//
// --
// --
// # Proposal #0 - Setting prop1.string param
//
// ## Description
//
// setting prop1.string param
//
// ## Proposal information
//
// **Status: EXECUTION SUCCESSFUL**
//
// **Voting stats:**
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTES 0 (0%)
//
//
// **Threshold met: TRUE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
// ### Actions
//
// The voting period for this proposal is over.
//
//
