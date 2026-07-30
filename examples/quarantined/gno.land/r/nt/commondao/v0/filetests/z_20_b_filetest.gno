// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_b_filetest

package z_20_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The execution kind is NOT seeded on new DAOs. Creating an execution
	// proposal before the DAO enables the kind is refused by the opt-in
	// enable-gate (assertKindEnabled). Removing that gate would let this
	// call reach Propose, which rejects the unregistered kind with a
	// different message ("proposal kind not found") — this golden pins the
	// gate.
	println("execution enabled:", commondao.IsProposalKindEnabled(daoID, "execution"))
	commondao.CreateExecutionProposal(cross(cur), daoID, "Run it", "Body", func(_ int, _ realm) error {
		return nil
	})
}

// Output:
// execution enabled: false

// Error:
// proposal kind is not enabled: execution
