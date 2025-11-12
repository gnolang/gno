package custom_resource

import (
	"testing"

	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
)

func TestDAOInitialization(t *testing.T) {
	t.Run("Test member initialization", func(t *testing.T) {
		membersCount := daoPrivate.Members.MembersCount()
		if membersCount != 4 {
			t.Fatalf("Expected 4 members, got %d", membersCount)
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

	t.Run("Test custom resource registration", func(t *testing.T) {
		res := daoPrivate.Core.Resources.Get(ActionNewPostKind)
		if res == nil {
			t.Fatalf("Expected '%s' resource to be registered", ActionNewPostKind)
		}
		if res.DisplayName != "Create Blog Post" {
			t.Errorf("Expected resource display name 'Create Blog Post', got '%s'", res.DisplayName)
		}
		if res.Description != "Allows DAO members to create and publish new blog posts" {
			t.Errorf("Allows DAO members to create and publish new blog posts', got '%s'", res.Description)
		}
	})
}

func TestBlogResource(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm("g1demo1234567890abcdefghijklmnopqrstuvw1"))

	action := NewPostAction("Test Title", "Test Content")
	// Create test proposal
	req := daokit.ProposalRequest{
		Title:       "Add new blog post",
		Description: "Adding an important announcement",
		Action:      action,
	}

	t.Run("Test proposal creation", func(t *testing.T) {
		Propose(cross, req)
		// Should be 4 total proposals (3 initial + 1 test proposal)
		if daoPrivate.Core.Proposals.Tree.Size() != 4 {
			t.Fatalf("Expected 4 total proposals, got %d", daoPrivate.Core.Proposals.Tree.Size())
		}

		prop := daoPrivate.Core.Proposals.GetProposal(4)
		if prop.Title != req.Title {
			t.Errorf("Expected title '%s', got '%s'", req.Title, prop.Title)
		}
	})

	t.Run("Test voting and execution", func(t *testing.T) {
		// Verify initial blog state
		if len(myblog.Post) != 0 {
			t.Fatalf("Expected 0 posts initially, got %d", len(myblog.Post))
		}

		// Vote with 1 member (meets 10% threshold of 4 members)
		Vote(cross, 4, daocond.VoteYes) // member1

		// Execute the proposal
		Execute(cross, 4)

		// Verify blog post was created
		if len(myblog.Post) != 1 {
			t.Fatalf("Expected 1 post after execution, got %d", len(myblog.Post))
		}

		post := myblog.Post[0]
		if post.Title != "Test Title" {
			t.Errorf("Expected title 'Test Title', got '%s'", post.Title)
		}
		if post.Content != "Test Content" {
			t.Errorf("Expected content 'Test Content', got '%s'", post.Content)
		}
	})

	t.Run("Test insufficient votes", func(t *testing.T) {
		// Reset blog
		myblog = Blog{}

		// Create new proposal
		req.Action = NewPostAction("Should Not Create", "This shouldn't appear")
		Propose(cross, req)

		// should panic
		abort := revive(func() { Execute(cross, 2) })
		if abort != "proposal condition is not met" {
			t.Fatalf("Expected execute to panic with 'proposal condition is not met', but got: '%s'", abort)
		}

		// Verify blog post was NOT created
		if len(myblog.Post) != 0 {
			t.Fatalf("Expected 0 posts after failed execution, got %d", len(myblog.Post))
		}
	})
}
