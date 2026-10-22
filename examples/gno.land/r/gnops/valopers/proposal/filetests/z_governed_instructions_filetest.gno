// PKGPATH: gno.land/r/gnops/valopers/proposal/filetests/z_governed_instructions_filetest
//
// Positive counterpart to
// r/gnops/valopers/filetests/z_foreign_realm_capability_filetest.gno.
//
// Sealing the privileged closure inside a dao.ProposalRequest closes the
// capability leak, but it would be worthless if it
// also broke the legitimate path. This runs the real thing end to end —
// propose, vote, execute — and asserts the instructions are actually
// rewritten.
//
// Read the two files together: same write, but reachable ONLY with a
// GovDAO majority behind it.
package z_governed_instructions_filetest

import (
	"chain/runtime/unsafe"
	"strings"
	"testing"

	"gno.land/r/gnops/valopers"
	"gno.land/r/gnops/valopers/proposal"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/init/v0"
)

var member address = unsafe.OriginCaller()

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	pr := proposal.ProposeNewInstructionsProposalRequest(cross(cur), "GOVERNED-INSTRUCTIONS")
	dao.MustCreateProposal(cross(cur), pr)
}

func main(cur realm) {
	testing.SetOriginCaller(member)

	println("before:", strings.Contains(valopers.Render(""), "GOVERNED-INSTRUCTIONS"))

	// Voter-facing disclosure (the pr6068 anti-phishing line, rendered as
	// "Executor created in: ..."). It moved from r/gnops/valopers/proposal
	// to r/gnops/valopers when request construction moved into the realm,
	// so pin it rather than let it drift unobserved.
	prop, err := dao.GetProposal(dao.ProposalID(0))
	if err != nil {
		panic(err)
	}
	println("title:", prop.Title())
	println("executor created in:", prop.ExecutorCreationRealm())

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed:", dao.ExecuteProposal(cross(cur), dao.ProposalID(0)))

	println("after:", strings.Contains(valopers.Render(""), "GOVERNED-INSTRUCTIONS"))
}

// Output:
// before: false
// title: /r/gnops/valopers: Update instructions
// executor created in: gno.land/r/gnops/valopers
// executed: true
// after: true
