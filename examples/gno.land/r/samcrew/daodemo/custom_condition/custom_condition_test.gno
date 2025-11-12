package custom_condition

import (
	"testing"

	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daocond"
)

func TestNoRoleCountCond(t *testing.T) {
	// Setup test members and roles
	initialRoles := []basedao.RoleInfo{
		{Name: "Role1", Description: "Test role 1", Color: "#329175"},
		{Name: "Role2", Description: "Test role 2", Color: "#21577A"},
	}

	initialMembers := []basedao.Member{
		{Address: "member1", Roles: []string{"Role1"}}, // has role
		{Address: "member2", Roles: []string{}},        // no role
		{Address: "member3", Roles: []string{}},        // no role
		{Address: "member4", Roles: []string{"Role2"}}, // has role
	}

	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)

	t.Run("Test hasNoRole helper", func(t *testing.T) {
		if !hasNoRole(memberStore, "member2") {
			t.Error("member2 should have no roles")
		}
		if hasNoRole(memberStore, "member1") {
			t.Error("member1 should have roles")
		}
	})

	t.Run("Test Eval with sufficient votes", func(t *testing.T) {
		cond := NoRole(memberStore, 2)
		ballot := daocond.NewBallot()
		ballot.Vote("member2", daocond.VoteYes) // no role
		ballot.Vote("member3", daocond.VoteYes) // no role

		if !cond.Eval(ballot) {
			t.Error("Condition should be satisfied with 2 roleless yes votes")
		}
	})

	t.Run("Test Eval with insufficient votes", func(t *testing.T) {
		cond := NoRole(memberStore, 2)
		ballot := daocond.NewBallot()
		ballot.Vote("member2", daocond.VoteYes) // no role
		ballot.Vote("member1", daocond.VoteYes) // has role - shouldn't count

		if cond.Eval(ballot) {
			t.Error("Condition should not be satisfied with only 1 roleless yes vote")
		}
	})

	t.Run("Test Signal calculation", func(t *testing.T) {
		cond := NoRole(memberStore, 2)
		ballot := daocond.NewBallot()
		ballot.Vote("member2", daocond.VoteYes) // no role

		signal := cond.Signal(ballot)
		if signal != 0.5 {
			t.Errorf("Expected signal 0.5, got %f", signal)
		}

		ballot.Vote("member3", daocond.VoteYes) // no role
		signal = cond.Signal(ballot)
		if signal != 1.0 {
			t.Errorf("Expected signal 1.0, got %f", signal)
		}
	})

	t.Run("Test Render methods", func(t *testing.T) {
		cond := NoRole(memberStore, 3)
		ballot := daocond.NewBallot()
		ballot.Vote("member2", daocond.VoteYes) // no role

		// Test basic render
		rendered := cond.Render()
		if rendered != "3" {
			t.Errorf("Expected render '3', got '%s'", rendered)
		}

		// Test render with votes
		renderedWithVotes := cond.RenderWithVotes(ballot)
		expected := "3 members with no role must vote yes\n\nYes: 1/3\n\n"
		if renderedWithVotes != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, renderedWithVotes)
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("Test member initialization", func(t *testing.T) {
		membersCount := daoPrivate.Members.MembersCount()
		if membersCount != 3 {
			t.Fatalf("Expected 3 members, got %d", membersCount)
		}
	})

	t.Run("Test role initialization", func(t *testing.T) {
		roles := daoPrivate.Members.GetRoles()
		expectedRoles := []string{"Role"}
		if len(roles) != len(expectedRoles) {
			t.Fatalf("Expected %d roles, got %d", len(expectedRoles), len(roles))
		}
		for i, role := range roles {
			if role != expectedRoles[i] {
				t.Fatalf("Expected role %s, got %s", expectedRoles[i], role)
			}
		}
	})
}
