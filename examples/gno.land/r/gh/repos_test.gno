package gh

import "testing"

// Test for the Repo struct functions.
func TestRepoFunctions(t *testing.T) {
	account := &Account{
		id:   "org123",
		name: "Sample Org",
		kind: "org",
	}
	repo := Repo{
		id:    "org123/sample-repo",
		owner: account,
		name:  "sample-repo",
	}

	t.Run("Test Repo ID", func(t *testing.T) {
		if repo.ID() != "org123/sample-repo" {
			t.Fatalf("Expected ID to be org123/sample-repo, got %s", repo.ID())
		}
	})

	t.Run("Test Repo Name", func(t *testing.T) {
		if repo.Name() != "sample-repo" {
			t.Fatalf("Expected Name to be sample-repo, got %s", repo.Name())
		}
	})

	t.Run("Test Repo Owner", func(t *testing.T) {
		if repo.Owner().ID() != "org123" {
			t.Fatalf("Expected Owner ID to be org123, got %s", repo.Owner().ID())
		}
	})

	t.Run("Test Repo URL", func(t *testing.T) {
		if repo.URL() != "https://github.com/org123/sample-repo" {
			t.Fatalf("Expected URL to be https://github.com/org123/sample-repo, got %s", repo.URL())
		}
	})
}
