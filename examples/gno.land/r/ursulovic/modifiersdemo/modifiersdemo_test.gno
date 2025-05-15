package modifiersdemo

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/ursulovic/functionmodifiers"
)

func TestInitializeOwners(t *testing.T) {

	// Test empty owners list (should panic)
	t.Run("empty_owners", func(t *testing.T) {
		owners := []string{}
		newOwnerPercentage := 0.5

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Expected function to panic with empty owners list, but it did not")
			}
		}()
		Modifiers.InitializeOwners(owners, newOwnerPercentage)
	})

	// Test invalid address (should panic)
	t.Run("invalid_address", func(t *testing.T) {
		owners := []string{"invalid!address"}
		newOwnerPercentage := 0.5

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Expected function to panic with invalid address, but it did not")
			}
		}()
		Modifiers.InitializeOwners(owners, newOwnerPercentage)
	})

	// Test valid initialization
	t.Run("valid_initialization", func(t *testing.T) {
		owners := []string{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
		newOwnerPercentage := 0.5

		Modifiers.InitializeOwners(owners, newOwnerPercentage)

		// Verify owner was added
		if !Modifiers.Owners.Has(owners[0]) {
			t.Fatal("Owner was not properly initialized")
		}
	})

	// Test negative percentage (should panic)
	t.Run("negative_percentage", func(t *testing.T) {
		owners := []string{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
		newOwnerPercentage := -0.5

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Expected function to panic with negative percentage, but it did not")
			}
		}()
		Modifiers.InitializeOwners(owners, newOwnerPercentage)
	})

}

func TestProposeNewOwner(t *testing.T) {
	// Setup initial owner
	initialOwner := "g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x"
	Modifiers = functionmodifiers.ModifierRegistry{} // Initialize the struct
	Modifiers.InitializeOwners([]string{initialOwner}, 0.5)

	// Set the caller to be the initial owner
	std.TestSetOrigCaller(std.Address(initialOwner))

	// Test proposing already existing owner (should panic)
	t.Run("existing_owner", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Expected function to panic when proposing existing owner, but it did not")
			}
		}()
		Modifiers.ProposeNewOwner(initialOwner)
	})

	// Test valid proposal
	t.Run("valid_proposal", func(t *testing.T) {
		newOwner := "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
		Modifiers.ProposeNewOwner(newOwner)

		// Verify proposal was added
		proposals := Modifiers.Proposals.OwnerProposals
		v, exists := proposals.Get(newOwner)
		if !exists {
			t.Fatal("New owner proposal was not properly added")
		}

		// Verify the approval was recorded
		approvals := v.(*avl.Tree)
		if !approvals.Has(initialOwner) {
			t.Fatal("Owner approval was not properly recorded")
		}
	})
}
