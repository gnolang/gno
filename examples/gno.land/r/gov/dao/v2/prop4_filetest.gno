package main

import (
	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
	_ "gno.land/r/gov/dao/init" // so that the govdao initializer is executed
	govdaov2 "gno.land/r/gov/dao/v2"
	"gno.land/r/sys/params"
)

func init() {
	mExec := params.NewSysParamStringPropExecutor("foo", "bar", "baz", "quz")
	title := "Setting foo:bar:baz param"
	comment := "new value will be quz"
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
// ## [Prop #0 - Setting foo:bar:baz param](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// # Proposal #0 - Setting foo:bar:baz param
//
// ## Description
//
// new value will be quz
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
// #### [[Vote YES](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=YES)] - [[Vote NO](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=NO)] - [[Vote ABSTAIN](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=ABSTAIN)]
//
//
// --
// --
// # Proposal #0 - Setting foo:bar:baz param
//
// ## Description
//
// new value will be quz
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
// # Proposal #0 - Setting foo:bar:baz param
//
// ## Description
//
// new value will be quz
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
