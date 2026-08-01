// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_21_a_filetest

// A DAO acting as a council member of another DAO.
//
// P (the "parent") has two council members: a person (`user`) and another DAO
// (`M`). A DAO has no private key, so M cannot sign a Vote transaction. A DAO
// votes by running an EXECUTION proposal in its own governance whose closure
// calls commondao.Vote as M's own sub-identity: the host mints M's sub
// (cur.Sub("dao/<M>")) and the closure crosses into Vote with it, so P sees
// M's address as the caller and records M's vote.
//
// The closure must be authored in a PERSISTENT realm — a closure from a CLI
// `maketx run` does not survive Propose->Execute. This filetest is itself a
// persistent realm, so it plays that role; on chain you would deploy a small
// "policy" realm that sits on M's council and authors the closure.
//
// Transaction sequence (each commondao.* / policy-realm call is one tx):
//   setup: create M; register the execution kind on M; create P with {user, M}
//   in P:  1. CreateTextProposal      (user)         -> propP
//          2. Vote YES on propP        (user)         [propP: 1/2, active]
//   in M:  3. CreateExecutionProposal  (policy realm) -> propM
//                closure := Vote YES on (P, propP) as M
//          4. Vote YES on propM        (M's council)
//          5. Execute propM            -> runs closure -> M votes YES in P
//                                                        [propP: 2/2, passed]
//   in P:  6. Execute propP            -> Executed
package z_21_a_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx (genesis council)
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1 (operator)
)

var (
	memberDAO uint64  // M
	parentDAO uint64  // P
	mAddr     address // M's on-chain address (a council member of P)
)

func init(cur realm) {
	// The genesis council invites the operator, who may then create DAOs.
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))

	// Create M, the DAO that will be a member of P. (Single-member council
	// here, standing in for however M is really governed.)
	memberDAO = commondao.New(cross(cur), "MemberDAO", "Acts as a member of another DAO", "", "")

	// A DAO can only act as its own address through the execution kind, which
	// is opt-in. Register it on M through M's own governance.
	pid := commondao.CreateRegisterKindProposal(cross(cur), memberDAO, "execution")
	commondao.Vote(cross(cur), memberDAO, pid, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), memberDAO, pid)

	// M's address is what we place on P's council.
	mAddr = commondao.GetView(memberDAO).Address()

	// Create P with two council members: the person `user` and the DAO `M`.
	parentDAO = commondao.New(cross(cur), "ParentDAO", "Governed by a person and a DAO", "", mAddr.String())
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	println("P has the member DAO on its council:", commondao.GetView(parentDAO).Council().Has(mAddr))

	// Tx 1 + 2: a proposal in P, and the human member's YES vote. One YES is
	// not enough: a two-member council needs both to reach the threshold.
	propP := commondao.CreateTextProposal(cross(cur), parentDAO, "Parent decision", "Should P do X?", 7)
	commondao.Vote(cross(cur), parentDAO, propP, pdao.ChoiceYes, "person votes yes")
	println("propP after the person's vote:", string(statusOf(parentDAO, propP)))

	// Tx 3: M creates an execution proposal. The closure is authored in this
	// persistent realm and captures P's IDs; at execution it receives M's sub
	// and votes as M.
	propM := commondao.CreateExecutionProposal(cross(cur), memberDAO,
		"Vote in ParentDAO", "Cast this DAO's YES vote on the parent proposal.",
		func(_ int, sub realm) error {
			commondao.Vote(cross(sub), parentDAO, propP, pdao.ChoiceYes, "cast as the member DAO")
			return nil
		},
	)

	// Tx 4 + 5: M's council passes the execution proposal, then executes it.
	// Executing runs the closure, which casts M's vote in P as M's own address.
	commondao.Vote(cross(cur), memberDAO, propM, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), memberDAO, propM)

	pP, _ := commondao.GetView(parentDAO).GetProposal(propP)
	println("M's vote recorded in P:", pP.VotingRecord().HasVoted(mAddr))
	println("propP YES votes:", pP.VotingRecord().VoteCount(pdao.ChoiceYes))
	println("propP after the member DAO's vote:", string(pP.Status()))

	// Tx 6: finalize the now-passed parent proposal.
	commondao.Execute(cross(cur), parentDAO, propP)
	println("propP final status:", string(statusOf(parentDAO, propP)))
}

func statusOf(daoID, propID uint64) pdao.ProposalStatus {
	p, _ := commondao.GetView(daoID).GetProposal(propID)
	return p.Status()
}

// Output:
// P has the member DAO on its council: true
// propP after the person's vote: active
// M's vote recorded in P: true
// propP YES votes: 2
// propP after the member DAO's vote: passed
// propP final status: executed
