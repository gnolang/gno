// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_a_filetest

package z_7_a_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
)

var parentID uint64

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(testing.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v0"))
	parentDAO := commondao.New(cross(cur), "Parent DAO", "Purpose", "", "")
	parentID = parentDAO
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v0"))

	// SubDAOs are created through a council proposal at simple majority
	pID := commondao.CreateSubDAOProposal(cross(cur), parentID, name, "Purpose", "Sub charter", string(user))
	commondao.Vote(cross(cur), parentID, pID, pcommondao.ChoiceYes, "")
	commondao.Execute(cross(cur), parentID, pID)

	subID := uint64(3)
	view := commondao.GetView(subID)

	println(view.Name() == name)
	println(view.Description())

	parent, found := view.Parent()
	println(found && parent.ID() == parentID)

	// Check that SubDAO is added as a child to the parent DAO
	childFound := false
	parent.IterateChildren(func(child pcommondao.ReadonlyCommonDAO) bool {
		childFound = child.ID() == subID
		return childFound
	})
	println(childFound)
}

// Output:
// true
// Sub charter
// true
// true
