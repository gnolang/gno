package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
	users "gno.land/r/sys/namereg/v1"
	susers "gno.land/r/sys/users"
)

// Test the full governance-driven pause/unpause flow:
//  1. propose pause via NewSetPausedExecutor(true), vote YES, execute,
//     observe IsPaused() flip to true.
//  2. propose unpause via NewSetPausedExecutor(false), vote YES, execute,
//     observe IsPaused() flip back to false.
//  3. Register works again post-unpause.
var c address = runtime.OriginCaller()

func init() {
	// Whitelist this realm as a controller so its Register() can reach r/sys/users.
	testing.SetHeight(0)
	susers.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	testing.SetHeight(123)

	daov3init.InitWithUsers(c)
}

func main() {
	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))

	println("paused before:", users.IsPaused())

	// Propose, vote, execute: pause.
	dao.MustCreateProposal(cross, users.NewSetPausedExecutor(true))
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})
	dao.ExecuteProposal(cross, dao.ProposalID(0))
	println("paused after pause prop:", users.IsPaused())

	// Propose, vote, execute: unpause.
	dao.MustCreateProposal(cross, users.NewSetPausedExecutor(false))
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(1),
	})
	dao.ExecuteProposal(cross, dao.ProposalID(1))
	println("paused after unpause prop:", users.IsPaused())

	// Register works again (the pause check inside Register would have
	// panicked with ErrPaused had the unpause executor not run).
	users.Register(cross, "nym-alice123")
	data, _ := susers.ResolveName("nym-alice123")
	println("registered alice at:", data.Addr())
}

// Output:
// paused before: false
// paused after pause prop: true
// paused after unpause prop: false
// registered alice at: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
