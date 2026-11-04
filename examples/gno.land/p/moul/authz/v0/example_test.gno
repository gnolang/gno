package authz

// Example_basic demonstrates initializing and using a basic member authority
func Example_basic(cur realm) {
	// Initialize from the EOA caller (e.g. in init(cur realm) of a realm
	// being deployed): caller passes the EOA address; the realm itself
	// is responsible for verifying the caller is an EOA when needed.
	auth := NewWithMembers(cur.Previous().Address())

	// Authorize with DoByPrevious, NOT DoByCurrent: the member set holds
	// cur.Previous().Address(), and DoByCurrent would present
	// cur.Address() -- this realm -- which is never in that set, so the
	// action would silently never run. Seeding and authorizing must read
	// the SAME side of the frame. Matches the package Quick Start.
	if err := auth.DoByPrevious(0, cur, "update_config", func() error {
		// config = newValue
		return nil
	}); err != nil {
		panic(err)
	}
}

// Example_addingMembers demonstrates how to add new members to a member authority
func Example_addingMembers(cur realm) {
	// Seed with Previous(), not Address(): MemberAuthority.AddMember
	// derives its principal as rlm.Previous().Address(), so a set holding
	// only cur.Address() authorizes nobody and the AddMember below fails.
	auth := NewWithMembers(cur.Previous().Address())

	// Add a new member to the authority
	memberAuth := auth.Authority().(*MemberAuthority)
	if err := memberAuth.AddMember(0, cur, address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")); err != nil {
		panic(err)
	}
}

// Example_contractAuthority demonstrates a contract-based authority.
//
// A plain NewContractAuthority accepts exactly one principal:
// chain.PackageAddress(path). Point `path` at the principal you want to
// ASSERT -- normally the governance realm that will drive the action --
// and authorize with DoByPrevious, so the comparison is against whoever
// crossed in. See NewContractAuthority's godoc.
//
// Do NOT pair a plain ContractAuthority on your OWN path with
// DoByCurrent. rlm.Address() inside any crossing frame of a realm is
// unconditionally PackageAddress(ownPath), so that gate compares the
// realm to itself, can never reject, and any exported crossing
// entrypoint next to it becomes an unauthenticated privileged write --
// the self-comparing-gate shape. Own-path is right only when something
// OTHER than your own frame supplies the caller, i.e. DoByPrevious from
// a realm you deliberately let call in, or Transfer.
//
// Nothing here catches a wrong pairing: Example_* functions that take
// any parameter are never executed by `gno test` (see isExampleFunc in
// gnovm/pkg/test/test.go -- it rejects on method receiver, parameters or
// results; "// Output:" plays no part in the predicate, so adding one
// does NOT make the body run), so a sentinel panic in here still reports
// ok. filetests/z_contract_authority_shape_filetest.gno executes the
// gate from a real realm instead.
//
// Corollary of the identity gate: do not export anything that returns a
// crossing closure of this realm, or callers inherit the authority.
func Example_contractAuthority(cur realm) {
	// This realm asserts "the governance realm governs me".
	auth := NewWithAuthority(
		NewContractAuthority(
			"gno.land/r/gov/dao", // the principal asserted, not this realm
			mockDAOHandler,       // defined elsewhere for example
		),
	)

	// Authorized only when gno.land/r/gov/dao is the realm that crossed
	// into this function -- in practice, when a passed proposal executes.
	// Every other caller presents its own address here and is refused,
	// which is the whole point of the gate. Never discard this error.
	if err := auth.DoByPrevious(0, cur, "update_params", func() error {
		return nil
	}); err != nil {
		panic(err)
	}
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

	// Only members can propose, and contract must approve.
	//
	// DoByPrevious, not DoByCurrent: the proposer set holds EOA
	// addresses, and DoByCurrent would present this realm's own address,
	// which is in no member set -- the proposer would reject and the
	// action would silently never run.
	if err := auth.DoByPrevious(0, cur, "update_params", func() error {
		// Executes after:
		// 1. Proposer initiates
		// 2. DAO approves
		return nil
	}); err != nil {
		panic(err)
	}
}

// Example_switchingAuthority demonstrates switching from member to contract
// authority.
//
// Note what a plain ContractAuthority on a FOREIGN path means after the
// transfer: this realm loses the ability to take the authority back, and
// DoByCurrent against it is dead. It is NOT a one-way door in the
// stronger sense an earlier draft implied — the foreign realm can still
// drive and rotate it via Previous() once it crosses in, which is the
// intended async-DAO shape. What is given up is the original owner's
// claim, not the authority's mutability. (The ADR for this change states
// the same thing; keep the two in step.) To keep the authority here
// while letting others propose, transfer to
// NewRestrictedContractAuthority(ownPath, handler, proposerAuth)
// instead.
func Example_switchingAuthority(cur realm) {
	// Start with member authority.
	//
	// Seed the member set with Previous(), not Address(): Transfer
	// derives its principal as rlm.Previous().Address(), so a set
	// containing only cur.Address() authorizes nobody but a
	// same-package cross-call, and the Transfer below would abort for
	// every external caller — including a realm that copies this into
	// init(cur realm), which then fails its deploy tx. Matches
	// Example_basic and NewWithMembers' godoc.
	auth := NewWithMembers(cur.Previous().Address())

	// Create and switch to contract authority — control moves to
	// gno.land/r/demo/dao.
	daoAuthority := NewContractAuthority(
		"gno.land/r/demo/dao",
		mockDAOHandler,
	)
	if err := auth.Transfer(0, cur, daoAuthority); err != nil {
		panic(err)
	}
}

// Mock handler for examples
func mockDAOHandler(title string, action PrivilegedAction) error {
	return action()
}
