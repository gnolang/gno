package authz

// Example_basic demonstrates initializing and using a basic member authority
func Example_basic(cur realm) {
	// Initialize from the EOA caller (e.g. in init(cur realm) of a realm
	// being deployed): caller passes the EOA address; the realm itself
	// is responsible for verifying the caller is an EOA when needed.
	auth := NewWithMembers(cur.Previous().Address())

	// Use the authority to perform a privileged action
	auth.Do(0, cur, "update_config", func() error {
		// config = newValue
		return nil
	})
}

// Example_addingMembers demonstrates how to add new members to a member authority
func Example_addingMembers(cur realm) {
	// Initialize with the calling realm as the initial authority
	auth := NewWithMembers(cur.Address())

	// Add a new member to the authority
	memberAuth := auth.Authority().(*MemberAuthority)
	memberAuth.AddMember(cur.Address(), address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
}

// Example_contractAuthority demonstrates using a contract-based authority
func Example_contractAuthority(cur realm) {
	// Initialize with contract authority (e.g., DAO)
	auth := NewWithAuthority(
		NewContractAuthority(
			"gno.land/r/demo/dao",
			mockDAOHandler, // defined elsewhere for example
		),
	)

	// Privileged actions will be handled by the contract
	auth.Do(0, cur, "update_params", func() error {
		// Executes after DAO approval
		return nil
	})
}

// Example_restrictedContractAuthority demonstrates a contract authority with member-only proposals
func Example_restrictedContractAuthority(cur realm) {
	// Initialize member authority for proposers
	proposerAuth := NewMemberAuthority(
		address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"), // admin1
		address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), // admin2
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
	auth.Do(0, cur, "update_params", func() error {
		// Executes after:
		// 1. Proposer initiates
		// 2. DAO approves
		return nil
	})
}

// Example_switchingAuthority demonstrates switching from member to contract authority
func Example_switchingAuthority(cur realm) {
	// Start with member authority (the calling realm)
	addr := cur.Address()
	auth := NewWithMembers(addr)

	// Create and switch to contract authority
	daoAuthority := NewContractAuthority(
		"gno.land/r/demo/dao",
		mockDAOHandler,
	)
	auth.Transfer(addr, daoAuthority)
}

// Mock handler for examples
func mockDAOHandler(title string, action PrivilegedAction) error {
	return action()
}
