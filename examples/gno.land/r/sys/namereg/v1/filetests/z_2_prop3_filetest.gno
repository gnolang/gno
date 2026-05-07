package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
	users "gno.land/r/sys/namereg/v1"
	susers "gno.land/r/sys/users"
)

// Test the full governance-driven registerPrice flow:
//  1. propose a new price via users.ProposeNewRegisterPrice
//  2. DAO accepts the proposal (vote YES, execute)
//  3. verify the executed callback updated registerPrice by exercising
//     Register with the new amount and watching it succeed (registration
//     under the old default of 0 ugnot would now panic with the new
//     price in the error message).
var c address = runtime.OriginCaller()

func init() {
	// Whitelist this realm as a controller so its Register() can reach r/sys/users.
	testing.SetHeight(0)
	susers.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	testing.SetHeight(123)

	daov3init.InitWithUsers(c)

	// Propose raising the price from 0 to 20 GNOT.
	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))
	pr := users.ProposeNewRegisterPrice(20_000_000)
	dao.MustCreateProposal(cross, pr)
}

func main() {
	testing.SetOriginCaller(c)

	// Vote and execute.
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})
	dao.ExecuteProposal(cross, dao.ProposalID(0))

	// New price should be active. Register a user with the new amount;
	// registration with the old default (0 ugnot) would panic.
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginSend(chain.Coins{{Denom: "ugnot", Amount: 20_000_000}})
	users.Register(cross, "nym-alice123")

	data, _ := susers.ResolveName("nym-alice123")
	println("registered alice at:", data.Addr())
}

// Output:
// registered alice at: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
