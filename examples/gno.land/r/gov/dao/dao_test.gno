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

	expected := "# GovDAO Proposals\n\nNo proposals."
	urequire.Equal(t, expected, out)

	var called bool
	ex := pproposal.NewExecutor(func() error {
		called = true
		return nil
	})

	std.TestSetOrigCaller(u1)
	pid := Propose("dummy proposal title", "dummy proposal description", ex)

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

	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Active

#### Votes: 

- YAY: 1

- NAY: 0

- Percent: 33

- Members: 3

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)

	std.TestSetOrigCaller(u2)
	urequire.PanicsWithMessage(t, msgWrongVotingValue, func() {
		VoteOnProposal(pid, "INCORRECT")
	})
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "NO")
	})

	out = Render("0")

	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Active

#### Votes: 

- YAY: 1

- NAY: 1

- Percent: 33

- Members: 3

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)

	std.TestSetOrigCaller(u3)
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "YES")
	})

	out = Render("0")
	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Accepted

#### Votes: 

- YAY: 2

- NAY: 1

- Percent: 66

- Members: 3

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)

	// Add a new member, so non-executed proposals will change the voting status
	u4 := testutils.TestAddress("u4")
	members = append(members, u4)

	out = Render("0")
	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Active

#### Votes: 

- YAY: 2

- NAY: 1

- Percent: 50

- Members: 4

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)

	std.TestSetOrigCaller(u4)
	urequire.NotPanics(t, func() {
		VoteOnProposal(pid, "YES")
	})

	out = Render("0")
	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Accepted

#### Votes: 

- YAY: 3

- NAY: 1

- Percent: 75

- Members: 4

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)

	ExecuteProposal(pid)
	urequire.True(t, called)

	out = Render("0")

	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Succeeded

#### Votes: 

- YAY: 3

- NAY: 1

- Percent: 75

- Members: 4

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)

	// Add a new member and try to vote an already executed proposal
	u5 := testutils.TestAddress("u5")
	members = append(members, u5)
	std.TestSetOrigCaller(u5)
	urequire.PanicsWithMessage(t, msgPropExecuted, func() {
		ExecuteProposal(pid)
	})

	// even if we added a new member the executed proposal is showing correctly the members that voted on it
	out = Render("0")

	expected = `# Prop #0 - dummy proposal title

## Description

dummy proposal description

## Proposal information

#### Status: Succeeded

#### Votes: 

- YAY: 3

- NAY: 1

- Percent: 75

- Members: 4

#### Author: g1w5c47h6lta047h6lta047h6lta047h6ly5kscr`

	urequire.Equal(t, expected, out)
}
