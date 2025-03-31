package gh

import (
	"testing"
)

// Test for the Account struct functions.
func TestAccountFunctions(t *testing.T) {
	account := Account{
		id:   "user123",
		name: "John Doe",
		kind: "user",
	}

	t.Run("Test Account ID", func(t *testing.T) {
		if account.ID() != "user123" {
			t.Fatalf("Expected ID to be user123, got %s", account.ID())
		}
	})

	t.Run("Test Account Name", func(t *testing.T) {
		if account.Name() != "John Doe" {
			t.Fatalf("Expected Name to be John Doe, got %s", account.Name())
		}
	})

	t.Run("Test Account Kind", func(t *testing.T) {
		if account.Kind() != "user" {
			t.Fatalf("Expected Kind to be user, got %s", account.Kind())
		}
	})

	t.Run("Test Account URL", func(t *testing.T) {
		if account.URL() != "https://github.com/user123" {
			t.Fatalf("Expected URL to be https://github.com/user123, got %s", account.URL())
		}
	})

	t.Run("Test Account IsUser", func(t *testing.T) {
		if !account.IsUser() {
			t.Fatal("Expected IsUser to be true")
		}
	})

	t.Run("Test Account IsOrg", func(t *testing.T) {
		if account.IsOrg() {
			t.Fatal("Expected IsOrg to be false")
		}
	})
}
