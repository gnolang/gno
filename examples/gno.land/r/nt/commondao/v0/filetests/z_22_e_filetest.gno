// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_22_e_filetest

// The mandates/ folder is reserved: the Constitution grants a council
// self-power over its Bylaws only, so a council amendment targeting a
// mandates path is rejected at proposal creation.
package z_22_e_filetest

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

	payload := commondao.AmendBylawsPayload(daoID, "mandates/spend.md", "Spend only by vote.")
	commondao.CreateAmendBylawsProposal(cross(cur), daoID, payload)
}

// Error:
// mandates are not council-amendable: they are set at creation or by an ancestor (ancestor amendment is not implemented yet)
