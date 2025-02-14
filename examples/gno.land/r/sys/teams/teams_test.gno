// teams_test.gno
package teams

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

var (
	teamAddr   = testutils.TestAddress("team1")
	memberAddr = testutils.TestAddress("member1")
)

func TestRegister(t *testing.T) {
	defer cleanupTeam()

	// Valid registration
	func() {
		caller := std.GetOrigCaller()
		defer std.TestSetOrigCaller(caller)

		// Setup valid realm context
		std.TestSetOrigCaller(teamAddr)
		std.TestSetRealm(std.NewCodeRealm("gno.land/r/team1/home"))

		team := Register()
		uassert.True(t, team != nil)
		uassert.True(t, team == Get(teamAddr))
	}()

	// Test duplicate registration panic
	uassert.PanicsWithMessage(t, ErrAlreadyRegistered.Error(), func() {
		std.TestSetOrigCaller(teamAddr)
		std.TestSetRealm(std.NewCodeRealm("gno.land/r/team1/home"))
		Register()
	})
}

func TestMemberManagement(t *testing.T) {
	setupTeam()
	defer cleanupTeam()

	// Test initial state
	uassert.True(t, !HasMember(teamAddr, memberAddr))
	uassert.Equal(t, 0, Size(teamAddr))

	// Add member
	asTeam(func() {
		updated := MyTeam().Set(memberAddr, DefaultPermissions{CanAddPackage: true})
		uassert.False(t, updated, "should be new")
	})

	uassert.True(t, HasMember(teamAddr, memberAddr), "should have member")
	uassert.Equal(t, 1, Size(teamAddr))

	// Test duplicate set
	asTeam(func() {
		updated := MyTeam().Set(memberAddr, DefaultPermissions{CanAddPackage: false})
		uassert.True(t, updated, "should be updated")
	})
}

func TestPermissions(t *testing.T) {
	setupTeam()
	defer cleanupTeam()

	asTeam(func() {
		MyTeam().Set(memberAddr, DefaultPermissions{CanAddPackage: true})
	})

	// Valid permission
	uassert.True(t, HasPermission(teamAddr, memberAddr, ActionAddPackage{}))

	// Invalid permission
	uassert.True(t, !HasPermission(teamAddr, memberAddr, "other_action"))

	// Team address has full access
	uassert.True(t, HasPermission(teamAddr, memberAddr, ActionAddPackage{}))
}

func TestPagination(t *testing.T) {
	setupTeam()
	defer cleanupTeam()

	var members map[string]std.Address
	asTeam(func() {
		team := MyTeam()
		for i := 1; i <= 5; i++ {
			name := ufmt.Sprintf("member%d", i)
			addr := testutils.TestAddress(name)
			members[name] = addr
			team.Set(addr, DefaultPermissions{})
		}
	})

	// Test page 0
	page := Get(teamAddr).Page(0, 2)
	uassert.Equal(t, 2, len(page))
	uassert.Equal(t, members["member1"], page[0].String())
	uassert.Equal(t, members["member2"], page[1].String())

	// Test page 2
	page = Get(teamAddr).Page(2, 2)
	uassert.Equal(t, len(page), 1)
	uassert.Equal(t, members["member5"], page[0].String())
}

func TestAccessController(t *testing.T) {
	setupTeam()
	defer cleanupTeam()

	// Setup custom access controller
	ac := &testAC{allowed: memberAddr}

	asTeam(func() {
		MyTeam().SetAccessController(ac)
		MyTeam().Register(memberAddr)

	})

	// Verify custom permissions
	uassert.True(t, HasPermission(teamAddr, memberAddr, ActionAddPackage{}), "member addr")
	uassert.True(t, !HasPermission(teamAddr, "other_member", ActionAddPackage{}), "other member")
}

// Helpers
type testAC struct{ allowed std.Address }

func (ac *testAC) Can(member std.Address, do Action) bool {
	return member == ac.allowed
}

func asTeam(fn func()) {
	orig := std.GetOrigCaller()
	std.TestSetOrigCaller(teamAddr)
	defer std.TestSetOrigCaller(orig)
	fn()
}

func setupTeam() {
	asTeam(func() {
		std.TestSetRealm(std.NewCodeRealm("gno.land/r/team1/home"))
		Register()
	})
}

func cleanupTeam() {
	teams.Remove(teamAddr.String())
}
