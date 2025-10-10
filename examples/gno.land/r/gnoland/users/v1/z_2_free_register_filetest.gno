package main

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	users "gno.land/r/gnoland/users/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
	susers "gno.land/r/sys/users"
)

var c address = runtime.OriginCaller()

const gUsersV1Path = "gno.land/r/gnoland/users/v1"

// Register a namespace for test user
func registerTestUser(addrs ...address) {
	testing.SetRealm(testing.NewCodeRealm(gUsersV1Path))
	for _, addr := range addrs {
		err := susers.RegisterUser(cross, addr.String()[1:], addr)
		if err != nil {
			panic(err.Error() + " : " + addr.String())
		}
	}
}

func init() {
	// Initialize DAO with test user
	daov3init.InitWithUsers(c)
	registerTestUser(c)

	// Propose to set register price to 0
	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))
	pr := users.ProposeNewRegisterPrice(0)
	dao.MustCreateProposal(cross, pr)
}

func main() {
	testing.SetOriginCaller(c)

	// Vote and execute the proposal to set price to 0
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	dao.ExecuteProposal(cross, dao.ProposalID(0))

	println("--")
	println("Register price set to 0")
	println("--")

	// Now test registration with zero price using a different address (no coins sent)
	bob := testutils.TestAddress("bob")
	testing.SetOriginCaller(bob)
	testing.SetRealm(testing.NewUserRealm(bob))
	users.Register(cross, "bob123")

	println("Successfully registered 'bob123' with zero registration price")
	println("--")
	println(users.Render(""))
}

// Output:
// --
// Register price set to 0
// --
// Successfully registered 'bob123' with zero registration price
// --
// # gno.land user registry
// Welcome to the gno.land user registry (v1). Please register a username.
//
// Registering a username grants the registering address the right to deploy packages and realms
// under that username’s namespace. For example, if an address registers the username `gnome123`, it
// will gain permission to deploy packages and realms to package paths with the pattern `gno.land/{p,r}/gnome123/*`.
//
// In V1, usernames must follow these rules, in order to prevent username squatting:
//
// - Must start with 3 letters
// - Letters must be lowercase
// - Must end with 3 numbers
// - Have a maximum length of 20 characters
// - With the only special character allowed being `_`
//
//
// In later versions of the registry, vanity usernames will be allowed through specific mechanisms.
//
// ###  [[Click here to register]]($help&func=Register)
// Registration price: 0.000000 GNOT (0ugnot)
//
// ---
//
//
// ## Latest registrations
// - [@bob123](/u/bob123)
