// Package authz provides flexible authorization control for privileged actions.
//
// # Authorization Strategies
//
// The package supports multiple authorization strategies:
//   - Member-based: Single user or team of users
//   - Contract-based: Async authorization (e.g., via DAO)
//   - Auto-accept: Allow all actions
//   - Drop: Deny all actions
//
// Core Components
//
//   - Authority interface: Base interface implemented by all authorities
//   - Authorizer: Main wrapper object for authority management
//   - MemberAuthority: Manages authorized addresses
//   - ContractAuthority: Delegates to another contract
//   - AutoAcceptAuthority: Accepts all actions
//   - DroppedAuthority: Denies all actions
//
// Quick Start
//
//	// Initialize with contract deployer as authority
//	var member address(...)
//	var auth = authz.NewWithMembers(member)
//
//	// Create functions that require authorization
//	func UpdateConfig(cur realm, newValue string) error {
//		return auth.DoByPrevious(0, cur, "update_config", func() error {
//			config = newValue
//			return nil
//		})
//	}
//
// See example_test.gno for more usage examples.
package authz

import (
	"chain"
	"errors"
	"strings"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/once"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/ufmt/v0"
)

// Authorizer is the main wrapper object that handles authority management.
// It is configured with a replaceable Authority implementation.
type Authorizer struct {
	auth Authority
}

// Authority represents an entity that can authorize privileged actions.
// It is implemented by MemberAuthority, ContractAuthority, AutoAcceptAuthority,
// and DroppedAuthority.
//
// Authority is the canonical safe shape for cross-package authority
// interfaces: methods are address-typed (no realm/cur crosses the interface
// boundary), and consumers correctly derive `caller` from
// `cur.Previous().Address()` under `rlm.IsCurrent()` before invoking
// Authorize. No cur-leak (class 1) is possible through this interface.
//
// However, two RESIDUAL RISKS apply:
//
//   - Class-3 impl-substitution: NewWithAuthority and Authorizer.Transfer
//     accept any Authority impl. A malicious Authority can always-approve
//     (silent privilege escalation) or always-deny (denial-of-service).
//     Consumers should pass canonical impls from this package
//     (MemberAuthority, ContractAuthority, AutoAcceptAuthority,
//     DroppedAuthority) unless they have explicit reason to register a
//     foreign impl. We do not expose an IsCanonicalAuthority allowlist
//     because the package is intentionally extensible — third-party impls
//     are the design intent.
//
//   - Class-4 closed-over-authority: NewContractAuthority and
//     NewRestrictedContractAuthority capture a caller-supplied
//     PrivilegedActionHandler closure. The handler runs synchronously
//     inside Authorize with the consumer's authority. A hostile handler
//     can swallow actions, log the caller, or execute arbitrary code
//     under the consumer's frame. Register only trusted handler functions.
//     See r/gnops/valopers/init.gno for the realistic registration shape.
//
// We do NOT seal Authority via an unexported marker method — that pattern
// is bypassable via embedding in Gno; see
// p/test/seal/filetests/z_seal_*_filetest.gno for the four bypass tests.
type Authority interface {
	// Authorize executes a privileged action if the caller is authorized
	// Additional args can be provided for context (e.g., for proposal creation)
	Authorize(caller address, title string, action PrivilegedAction, args ...any) error

	// String returns a human-readable description of the authority
	String() string
}

// PrivilegedAction defines a function that performs a privileged action.
type PrivilegedAction func() error

// PrivilegedActionHandler is called by contract-based authorities to handle
// privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// NewWithMembers creates a new Authorizer whose authority is a
// MemberAuthority containing the given addresses. Callers express
// authority intent at the call site:
//
//	// "auth realm is the authority"
//	a := authz.NewWithMembers(cur.Address())
//
//	// "previous realm is the authority" (from a crossing function)
//	a := authz.NewWithMembers(cur.Previous().Address())
//
//	// "EOA caller is the authority" (from init(cur realm))
//	if !cur.Previous().IsUserCall() {
//	    panic("realm must be initialized by EOA")
//	}
//	a := authz.NewWithMembers(cur.Previous().Address())
//
// This replaces the previous NewWithCurrent / NewWithPrevious /
// NewWithOrigin sugar — those baked runtime.{Current,Previous,Origin}
// reads into the constructor, which (a) prevented use from package-
// level var initializers, (b) made the EOA-origin check inside
// NewWithOrigin an indirect address comparison rather than the
// straightforward IsUserCall predicate, and (c) coupled the
// constructor to the runtime walks the rest of the migration is
// moving away from.
func NewWithMembers(addrs ...address) *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(addrs...),
	}
}

// NewWithAuthority creates a new Authorizer with a specific authority.
//
// SECURITY: `authority` is an open-interface input — any value satisfying
// Authority is accepted. A malicious impl can always-approve (privilege
// escalation) or always-deny (DoS). Prefer canonical impls from this
// package (NewMemberAuthority, NewContractAuthority, NewAutoAcceptAuthority,
// NewDroppedAuthority) unless you specifically need a foreign impl.
func NewWithAuthority(authority Authority) *Authorizer {
	return &Authorizer{
		auth: authority,
	}
}

// Authority returns the auth authority implementation
func (a *Authorizer) Authority() Authority {
	return a.auth
}

// Transfer changes the auth authority after validation. rlm must be the
// caller's own captured cur (asserted via rlm.IsCurrent()); the
// principal is rlm.Previous().Address(). Closes the address-parameter
// forgery: an external realm cannot supply Owner() as `caller` to
// bypass the underlying Authority's check.
//
// SECURITY (runtime substitution): once the current authority approves a
// Transfer, the new authority is installed and effective on the next call.
// If an attacker ever becomes the authority — even briefly — they can
// install a permanent DroppedAuthority (DoS) or an AutoAcceptAuthority
// (privilege escalation). Consumers concerned about this should wrap
// Transfer with a one-shot guard or a quorum/cooldown check.
//
// `newAuthority` is also an open-interface input — see NewWithAuthority's
// Class-3 caveat. Pass canonical impls.
func (a *Authorizer) Transfer(_ int, rlm realm, newAuthority Authority) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.auth.Authorize(caller, "transfer_authority", func() error {
		a.auth = newAuthority
		return nil
	})
}

// DoByCurrent executes a privileged action authorized as `rlm`. `rlm`
// must be the caller's own live cur (asserted via rlm.IsCurrent());
// the authorized principal is `rlm.Address()`. To authorize as the
// realm that called your function, use `DoByPrevious`.
//
//	auth.DoByCurrent(0, cur, "update_config", func() error { ... })    // current realm authorizes
//	auth.DoByPrevious(0, cur, "update_config", func() error { ... })   // calling realm authorizes
//
// The `_ int` first parameter is a deliberate sentinel that pushes
// `rlm realm` past the first-arg position so DoByCurrent stays a
// non-crossing method — otherwise it would be a crossing method and
// rlm.Previous() inside would resolve one realm deeper than the caller
// intended.
//
// SECURITY: the IsCurrent guard closes Class-2 designation forgery (see
// docs/resources/gno-security.md). A realm value's .Address() is set
// when the value is minted at a crossing frame; the value can in
// principle be stored and replayed. Without IsCurrent, a hostile realm
// could capture a high-privilege realm's cur.Previous() (e.g., when
// that realm called into it) and later pass the stored value here to
// authorize actions as that realm. IsCurrent rejects stale captures by
// requiring the value to match the topmost live crossing frame's cur.
func (a *Authorizer) DoByCurrent(_ int, rlm realm, title string, action PrivilegedAction, args ...any) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	return a.auth.Authorize(rlm.Address(), title, action, args...)
}

// DoByPrevious executes a privileged action authorized as the realm
// that called the function invoking DoByPrevious. `rlm` must be the
// caller's own live cur; the principal is derived as
// `rlm.Previous().Address()`. Mirrors the Transfer/AddMember pattern:
// always take live cur, derive the caller-of-caller internally rather
// than accepting a stored/forwarded realm value.
func (a *Authorizer) DoByPrevious(_ int, rlm realm, title string, action PrivilegedAction, args ...any) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	return a.auth.Authorize(rlm.Previous().Address(), title, action, args...)
}

// String returns a string representation of the auth authority
func (a *Authorizer) String() string {
	authStr := a.auth.String()

	switch a.auth.(type) {
	case *MemberAuthority:
	case *ContractAuthority:
	case *AutoAcceptAuthority:
	case *droppedAuthority:
	default:
		// this way official "dropped" is different from "*custom*: dropped" (autoclaimed).
		return ufmt.Sprintf("custom_authority[%s]", authStr)
	}
	return authStr
}

// MemberAuthority is the default implementation using addrset for member
// management.
type MemberAuthority struct {
	members addrset.Set
}

func NewMemberAuthority(members ...address) *MemberAuthority {
	auth := &MemberAuthority{}
	for _, addr := range members {
		auth.members.Add(addr)
	}
	return auth
}

func (a *MemberAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	if !a.members.Has(caller) {
		return errors.New("unauthorized")
	}

	if err := action(); err != nil {
		return err
	}
	return nil
}

func (a *MemberAuthority) String() string {
	addrs := []string{}
	a.members.Tree().Iterate("", "", func(key string, _ any) bool {
		addrs = append(addrs, key)
		return false
	})
	addrsStr := strings.Join(addrs, ",")
	return ufmt.Sprintf("member_authority[%s]", addrsStr)
}

// AddMember adds a new member to the authority. rlm must be the caller's
// own captured cur; the principal is rlm.Previous().Address() and must
// already be a member. The IsCurrent guard closes the forgery where an
// external realm passes Owner() as caller to bypass members.Has(caller).
func (a *MemberAuthority) AddMember(_ int, rlm realm, addr address) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.Authorize(caller, "add_member", func() error {
		a.members.Add(addr)
		return nil
	})
}

// AddMembers adds a list of members to the authority. Same rlm contract
// as AddMember.
func (a *MemberAuthority) AddMembers(_ int, rlm realm, addrs ...address) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.Authorize(caller, "add_members", func() error {
		for _, addr := range addrs {
			a.members.Add(addr)
		}
		return nil
	})
}

// RemoveMember removes a member from the authority. Same rlm contract
// as AddMember.
func (a *MemberAuthority) RemoveMember(_ int, rlm realm, addr address) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.Authorize(caller, "remove_member", func() error {
		a.members.Remove(addr)
		return nil
	})
}

// Tree returns a read-only view of the members tree
func (a *MemberAuthority) Tree() *rotree.ReadOnlyTree {
	tree := a.members.Tree().(*avl.Tree)
	return rotree.Wrap(tree, nil)
}

// Has checks if the given address is a member of the authority
func (a *MemberAuthority) Has(addr address) bool {
	return a.members.Has(addr)
}

// ContractAuthority implements async contract-based authority
type ContractAuthority struct {
	contractPath    string
	contractAddr    address
	contractHandler PrivilegedActionHandler
	proposer        Authority // controls who can create proposals
}

// NewContractAuthority creates a new contract-based authority.
//
// SECURITY (Class-4 captured callback): `handler` is a caller-supplied
// closure that runs SYNCHRONOUSLY inside Authorize, with the consumer's
// authority. A hostile handler can swallow actions, log the caller, or
// execute arbitrary code under the consumer's frame. The package-internal
// wrappedAction guards "execute action only from contractAddr" but the
// handler can call wrappedAction however it likes (multiple times, never,
// out of order). Register only trusted handler functions; treat handler
// registration as the trust boundary.
func NewContractAuthority(path string, handler PrivilegedActionHandler) *ContractAuthority {
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    chain.PackageAddress(path),
		contractHandler: handler,
		proposer:        NewAutoAcceptAuthority(), // default: anyone can propose
	}
}

// NewRestrictedContractAuthority creates a new contract authority with a
// proposer restriction.
//
// SECURITY:
//   - `handler` is the same Class-4 captured-callback risk as
//     NewContractAuthority — runs synchronously inside Authorize with the
//     consumer's authority. Register only trusted handler functions.
//   - `proposer` is an open-interface input (Class-3 impl-substitution).
//     A hostile proposer Authority can always-approve creation of any
//     proposal, defeating the restriction. Pass canonical impls only.
func NewRestrictedContractAuthority(path string, handler PrivilegedActionHandler, proposer Authority) Authority {
	if path == "" {
		panic("contract path cannot be empty")
	}
	if handler == nil {
		panic("contract handler cannot be nil")
	}
	if proposer == nil {
		panic("proposer cannot be nil")
	}
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    chain.PackageAddress(path),
		contractHandler: handler,
		proposer:        proposer,
	}
}

func (a *ContractAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	if a.contractHandler == nil {
		return errors.New("contract handler is not set")
	}

	// setup a once instance to ensure the action is executed only once
	executionOnce := once.Once{}

	// wrappedAction enforces at-most-once invocation. The previous
	// gate `unsafe.CurrentRealm() == contractAddr` is removed: it
	// was .Title()-bypassable (runtime.CurrentRealm walks past
	// non-crossing frames to the most-recent crossing ancestor) and
	// the trust boundary is now upstream — Authorizer.DoByCurrent /
	// DoByPrevious require rlm.IsCurrent() and pass a non-forgeable
	// principal to Authorize, while the consumer realm's handler
	// closure is the Class-4 trust root by lexical capture at
	// registration time.
	wrappedAction := func() error {
		return executionOnce.DoErr(func() error {
			return action()
		})
	}

	// Use the proposer authority to control who can create proposals
	return a.proposer.Authorize(caller, title+"_proposal", func() error {
		if err := a.contractHandler(title, wrappedAction); err != nil {
			return err
		}
		return nil
	}, args...)
}

func (a *ContractAuthority) String() string {
	return ufmt.Sprintf("contract_authority[contract=%s]", a.contractPath)
}

// AutoAcceptAuthority implements an authority that accepts all actions
// AutoAcceptAuthority is a simple authority that automatically accepts all
// actions.
// It can be used as a proposer authority to allow anyone to create proposals.
type AutoAcceptAuthority struct{}

func NewAutoAcceptAuthority() *AutoAcceptAuthority {
	return &AutoAcceptAuthority{}
}

func (a *AutoAcceptAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	return action()
}

func (a *AutoAcceptAuthority) String() string {
	return "auto_accept_authority"
}

// droppedAuthority implements an authority that denies all actions
type droppedAuthority struct{}

func NewDroppedAuthority() Authority {
	return &droppedAuthority{}
}

func (a *droppedAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	return errors.New("dropped authority: all actions are denied")
}

func (a *droppedAuthority) String() string {
	return "dropped_authority"
}
