package authz

import (
	"std"
)

// Example_basic demonstrates initializing and using a basic member authority
func Example_basic() {
	// Initialize with contract deployer as authority
	auth := NewWithOrigin()

	// Use the authority to perform a privileged action
	auth.Do("update_config", func() error {
		// config = newValue
		return nil
	})
}

// Example_addingMembers demonstrates how to add new members to a member authority
func Example_addingMembers() {
	// Initialize with contract deployer as authority
	auth := NewWithCurrent()

	// Add a new member to the authority
	memberAuth := auth.Current().(*MemberAuthority)
	memberAuth.AddMember(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
}

// Example_contractAuthority demonstrates using a contract-based authority
func Example_contractAuthority() {
	// Initialize with contract authority (e.g., DAO)
	auth := NewWithAuthority(
		NewContractAuthority(
			"gno.land/r/demo/dao",
			mockDAOHandler, // defined elsewhere for example
		),
	)

	// Privileged actions will be handled by the contract
	auth.Do("update_params", func() error {
		// Executes after DAO approval
		return nil
	})
}

// Example_restrictedContractAuthority demonstrates a contract authority with member-only proposals
func Example_restrictedContractAuthority() {
	// Initialize member authority for proposers
	proposerAuth := NewMemberAuthority(
		std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"), // admin1
		std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), // admin2
	)

	// Create contract authority with restricted proposers
	auth := NewWithAuthority(
		NewRestrictedContractAuthority(
			"gno.land/r/demo/dao",
			mockDAOHandler,
			proposerAuth,
		),
	)

	// Only members can propose, and contract must approve
	auth.Do("update_params", func() error {
		// Executes after:
		// 1. Proposer initiates
		// 2. DAO approves
		return nil
	})
}

// Example_switchingAuthority demonstrates switching from member to contract authority
func Example_switchingAuthority() {
	// Start with member authority (deployer)
	auth := NewWithCurrent()

	// Create and switch to contract authority
	daoAuthority := NewContractAuthority(
		"gno.land/r/demo/dao",
		mockDAOHandler,
	)
	auth.Transfer(daoAuthority)
}

// Mock handler for examples
func mockDAOHandler(title string, action PrivilegedAction) error {
	return action()
}
