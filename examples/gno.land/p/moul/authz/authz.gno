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
//	func UpdateConfig(newValue string) error {
//		crossing()
//		return auth.DoByPrevious("update_config", func() error {
//			config = newValue
//			return nil
//		})
//	}
//
// See example_test.gno for more usage examples.
package authz

import (
	"chain"
	"chain/runtime"
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

// Do executes a privileged action authorized as `rlm`. The caller
// chooses which realm authorizes by passing either their own `cur`
// (was DoByCurrent) or `cur.Previous()` (was DoByPrevious):
//
//	auth.Do(0, cur, "update_config", func() error { ... })            // current realm authorizes
//	auth.Do(0, cur.Previous(), "update_config", func() error { ... }) // calling realm authorizes
//
// The `_ int` first parameter is a deliberate sentinel that pushes
// `rlm realm` past the first-arg position so Do stays a non-crossing
// method — otherwise it would be a crossing method and rlm.Previous()
// inside would resolve one realm deeper than the caller intended.
func (a *Authorizer) Do(_ int, rlm realm, title string, action PrivilegedAction, args ...any) error {
	return a.auth.Authorize(rlm.Address(), title, action, args...)
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

	// Wrap the action to ensure it can only be executed by the contract
	wrappedAction := func() error {
		current := runtime.CurrentRealm().Address()
		if current != a.contractAddr {
			return errors.New("action can only be executed by the contract")
		}
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
