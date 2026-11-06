package teams

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const (
	aibHome   = "gno.land/r/aib/home"
	otherHome = "gno.land/r/other/home"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carol = testutils.TestAddress("carol")
)

// --- the caller-path parser -------------------------------------------------

func TestNamespaceOfHome(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		wantNS  string
		wantOK  bool
	}{
		{"canonical", "gno.land/r/aib/home", "aib", true},
		{"address namespace", "gno.land/r/g1abc/home", "g1abc", true},
		{"empty path", "", "", false},
		{"user realm has no path", "", "", false},
		{"pure package", "gno.land/p/aib/home", "", false},
		{"run realm", "gno.land/e/g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj/run", "", false},
		{"not home", "gno.land/r/aib/dao", "", false},
		{"nested home", "gno.land/r/aib/x/home", "", false},
		{"sub-realm identity", "gno.land/r/aib/home#admin", "", false},
		// cur.Sub() synthesizes "host#subpath". A subpath chosen to end in
		// /home makes the suffix check pass, so the "#" guard is the only
		// thing refusing these two.
		{"sub-realm ending in /home", "gno.land/r/evil#/home", "", false},
		{"sub-realm traversing to home", "gno.land/r/evil#../aib/home", "", false},
		{"sub-realm of nothing", "gno.land/r/home", "", false},
		{"empty namespace", "gno.land//home", "", false},
		{"home is the namespace", "gno.land/r/home/home", "home", true},
		{"suffix only", "home", "", false},
		{"wrong domain", "example.com/r/aib/home", "", false},
		{"prefix lookalike", "gno.land/rx/aib/home", "", false},
		{"trailing slash", "gno.land/r/aib/home/", "", false},
	}
	for _, tc := range cases {
		ns, ok := namespaceOfHome(tc.path)
		uassert.Equal(t, tc.wantOK, ok, tc.name+": ok")
		uassert.Equal(t, tc.wantNS, ns, tc.name+": ns")
	}
}

// --- the gate ---------------------------------------------------------------

func TestUserCallIsRefused(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	urequire.AbortsWithMessage(t, cur,
		"unauthorized: no calling realm; r/<ns>/home must call this",
		func() { AddMember(cross(cur), bob) },
	)
}

func TestNonHomeRealmIsRefused(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/aib/dao"))
	urequire.AbortsWithMessage(t, cur,
		"unauthorized: caller is not a <ns>/home realm: gno.land/r/aib/dao",
		func() { AddMember(cross(cur), bob) },
	)
}

// The sub-realm identity cur.Sub() synthesizes ("host#subpath") cannot be
// constructed here: testing.NewCodeRealm refuses it with "should only be called
// for Realms". TestNamespaceOfHome covers that shape at the parser instead,
// which is where the refusal actually lives.

// TestAssertHomeRejectsANonLiveRealm covers the IsCurrent check.
//
// This is the only shape where it can fire. assertHome takes its realm as an
// ordinary parameter, so an internal caller could hand it a stashed or
// sibling-frame value; without IsCurrent it would then derive the namespace
// from that value's Previous() and hand authority to whoever assembled it.
//
// The exported crossing functions do not need this test and cannot have it:
// the compiler refuses anything but cur or cross(rlm) as their first argument,
// so their cur is live by construction. Technique copied from
// r/sys/params TestAssertDelegateRejectsANonLiveRealm.
func TestAssertHomeRejectsANonLiveRealm(cur realm, t *testing.T) {
	fake := testing.MakeRealm(
		testutils.TestAddress("impostor"), "gno.land/r/impostor",
		testing.MakeRealm(testutils.TestAddress("h"), aibHome, testing.OriginRealm()),
	)
	uassert.Equal(t, aibHome, fake.Previous().PkgPath(),
		"premise: the forged realm must pass the path check, or this proves nothing")

	// PanicsContains, not Aborts: assertHome is called directly here, in this
	// realm, so the refusal is a panic rather than an abort.
	uassert.PanicsContains(t, cur, "not the caller's live cur", func() {
		assertHome(0, fake)
	})
}

func TestHomeWritesOnlyItsOwnNamespace(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(aibHome))
	AddMember(cross(cur), alice)

	testing.SetRealm(testing.NewCodeRealm(otherHome))
	AddMember(cross(cur), bob)

	uassert.True(t, IsMember("aib", alice), "alice in aib")
	uassert.False(t, IsMember("aib", bob), "bob not in aib")
	uassert.True(t, IsMember("other", bob), "bob in other")
	uassert.False(t, IsMember("other", alice), "alice not in other")
}

// --- membership -------------------------------------------------------------

func TestAddRemoveMember(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/add/home"))

	uassert.True(t, AddMember(cross(cur), alice), "first add reports true")
	uassert.False(t, AddMember(cross(cur), alice), "second add reports false")
	uassert.True(t, IsMember("add", alice), "alice is a member")

	uassert.True(t, RemoveMember(cross(cur), alice), "remove reports true")
	uassert.False(t, RemoveMember(cross(cur), alice), "second remove reports false")
	uassert.False(t, IsMember("add", alice), "alice is gone")
}

func TestSetMembersReplaces(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/set/home"))

	SetMembers(cross(cur), []address{alice, bob})
	uassert.Equal(t, 2, len(MembersOf("set", 0, MaxMembers)), "two members")

	SetMembers(cross(cur), []address{bob, carol})
	uassert.False(t, IsMember("set", alice), "alice dropped")
	uassert.True(t, IsMember("set", bob), "bob kept")
	uassert.True(t, IsMember("set", carol), "carol added")
	uassert.Equal(t, 2, len(MembersOf("set", 0, MaxMembers)), "still two members")
}

func TestSeedMembersOnceIsIdempotent(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/seed/home"))

	uassert.True(t, SeedMembersOnce(cross(cur), []address{alice}), "first seed writes")
	AddMember(cross(cur), bob)

	// This is what a redeploy does: init() runs again over existing state.
	uassert.False(t, SeedMembersOnce(cross(cur), []address{alice}), "re-seed is a no-op")
	uassert.True(t, IsMember("seed", bob), "a redeploy must not drop bob")
}

func TestInvalidMemberIsRefused(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/bad/home"))
	urequire.AbortsContains(t, cur, "invalid member address",
		func() { AddMember(cross(cur), address("nope")) },
	)
}

// --- retiring the registered-name path --------------------------------------

func TestRetireRequiresAMember(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/empty/home"))
	urequire.AbortsWithMessage(t, cur,
		"refused: retiring with an empty team would strand the namespace",
		func() { Retire(cross(cur)) },
	)
	uassert.False(t, IsRetired("empty"), "still not retired")
}

func TestRetireIsIdempotent(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/ret/home"))
	AddMember(cross(cur), alice)
	Retire(cross(cur))
	Retire(cross(cur))
	uassert.True(t, IsRetired("ret"), "retired")
}

func TestRetiredNamespaceCannotBeStranded(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/strand/home"))
	AddMember(cross(cur), alice)
	Retire(cross(cur))

	urequire.AbortsWithMessage(t, cur,
		"refused: removing the last member of a retired namespace would strand it",
		func() { RemoveMember(cross(cur), alice) },
	)
	urequire.AbortsWithMessage(t, cur,
		"refused: emptying the team of a retired namespace would strand it",
		func() { SetMembers(cross(cur), []address{}) },
	)
	uassert.True(t, IsMember("strand", alice), "alice is still there")
}

func TestUnknownNamespaceReadsFalse(t *testing.T) {
	uassert.False(t, HasTeam("never"), "no record")
	uassert.False(t, IsMember("never", alice), "not a member")
	uassert.False(t, IsRetired("never"), "not retired")
	uassert.Equal(t, 0, len(MembersOf("never", 0, MaxMembers)), "no members")
}

func TestMemberCapIsEnforced(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/cap/home"))

	big := make([]address, MaxMembers+1)
	for i := range big {
		big[i] = testutils.TestAddress(ufmt.Sprintf("m%d", i))
	}
	urequire.AbortsContains(t, cur, "exceeds the 128 cap",
		func() { SetMembers(cross(cur), big) },
	)

	SetMembers(cross(cur), big[:MaxMembers])
	uassert.Equal(t, MaxMembers, MemberCount("cap"), "at the cap")
	urequire.AbortsContains(t, cur, "at the 128 member cap",
		func() { AddMember(cross(cur), big[MaxMembers]) },
	)
	// Re-adding an existing member at the cap is not a growth and must pass.
	uassert.False(t, AddMember(cross(cur), big[0]), "already a member")
}

func TestRenderShowsStateAndRetirement(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/shown/home"))
	AddMember(cross(cur), alice)

	uassert.True(t, strings.Contains(Render("shown"), alice.String()), "member is listed")
	uassert.True(t, strings.Contains(Render("shown"), "path is live"), "not retired yet")
	uassert.True(t, strings.Contains(Render(""), "shown"), "index lists it")

	Retire(cross(cur))
	uassert.True(t, strings.Contains(Render("shown"), "**retired**"), "retirement is visible")
	uassert.True(t, strings.Contains(Render("absent"), "No team"), "unknown namespace")
}

func TestHasTeamIsFalseAfterTheLastMemberLeaves(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/hollow/home"))

	AddMember(cross(cur), alice)
	uassert.True(t, HasTeam("hollow"), "a team with a member")

	RemoveMember(cross(cur), alice)
	uassert.False(t, HasTeam("hollow"),
		"an emptied record is not a team: HasTeam must agree with IsMember")
	uassert.False(t, IsMember("hollow", alice), "and nobody is authorized")
}
