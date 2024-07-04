package govdao

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	pproposal "gno.land/p/gov/proposal"
)

func TestPackage(t *testing.T) {
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")
	u3 := testutils.TestAddress("u3")

	members = append(members, u1)
	members = append(members, u2)
	members = append(members, u3)

	nu1 := testutils.TestAddress("random1")

	out := Render("")

	expected := "No proposals found :(."
	urequire.Equal(t, expected, out)

	var called bool
	ex := pproposal.NewExecutor(func() error {
		called = true
		return nil
	})

	std.TestSetOrigCaller(u1)
	pid := Propose("dummy proposal", ex)

	// try to vote not being a member
	std.TestSetOrigCaller(nu1)

	urequire.PanicsWithMessage(t, msgCallerNotAMember, func() {
		VoteOnProposal(pid, "YES")
	})

	// try to vote several times
	std.TestSetOrigCaller(u1)
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "YES")
	})
	urequire.PanicsWithMessage(t, msgAlreadyVoted, func() {
		VoteOnProposal(pid, "YES")
	})

	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: active
Voting status: YES: 1, NO: 0, percent: 33, members: 3
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}

	std.TestSetOrigCaller(u2)
	urequire.PanicsWithMessage(t, msgWrongVotingValue, func() {
		VoteOnProposal(pid, "INCORRECT")
	})
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "NO")
	})

	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: active
Voting status: YES: 1, NO: 1, percent: 33, members: 3
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}

	std.TestSetOrigCaller(u3)
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "YES")
	})

	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: accepted
Voting status: YES: 2, NO: 1, percent: 66, members: 3
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}

	// Add a new member, so non-executed proposals will change the voting status
	u4 := testutils.TestAddress("u4")
	members = append(members, u4)

	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: active
Voting status: YES: 2, NO: 1, percent: 50, members: 4
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}

	std.TestSetOrigCaller(u4)
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "YES")
	})

	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: accepted
Voting status: YES: 3, NO: 1, percent: 75, members: 4
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}

	ExecuteProposal(pid)
	urequire.True(t, called)

	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: succeeded
Voting status: YES: 3, NO: 1, percent: 75, members: 4
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}

	// Add a new member and try to vote an already executed proposal
	u5 := testutils.TestAddress("u5")
	members = append(members, u5)
	std.TestSetOrigCaller(u5)
	urequire.PanicsWithMessage(t, msgPropExecuted, func() {
		ExecuteProposal(pid)
	})

	// even if we added a new member the executed proposal is showing correctly the members that voted on it
	out = Render("0")
	expected = `# Prop#0

dummy proposal
Status: succeeded
Voting status: YES: 3, NO: 1, percent: 75, members: 4
Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	if out != expected { // XXX: urequire.Equal is doing strange things with strings
		t.Errorf("%s != %s", out, expected)
	}
}
