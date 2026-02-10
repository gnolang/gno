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
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/rotree"
	"gno.land/p/nt/ufmt"
)

// Authorizer is the main wrapper object that handles authority management.
// It is configured with a replaceable Authority implementation.
type Authorizer struct {
	auth Authority
}

// Authority represents an entity that can authorize privileged actions.
// It is implemented by MemberAuthority, ContractAuthority, AutoAcceptAuthority,
// and DroppedAuthority.
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

// NewWithCurrent creates a new Authorizer with the auth realm's address as authority
func NewWithCurrent() *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(runtime.CurrentRealm().Address()),
	}
}

// NewWithPrevious creates a new Authorizer with the previous realm's address as authority
func NewWithPrevious() *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(runtime.PreviousRealm().Address()),
	}
}

// NewWithCurrent creates a new Authorizer with the auth realm's address as authority
func NewWithMembers(addrs ...address) *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(addrs...),
	}
}

// NewWithOrigin creates a new Authorizer with the origin caller's address as
// authority.
// This is typically used in the init() function.
func NewWithOrigin() *Authorizer {
	origin := runtime.OriginCaller()
	previous := runtime.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where runtime.PreviousRealm() is origin")
	}
	return &Authorizer{
		auth: NewMemberAuthority(origin),
	}
}

// NewWithAuthority creates a new Authorizer with a specific authority
func NewWithAuthority(authority Authority) *Authorizer {
	return &Authorizer{
		auth: authority,
	}
}

// Authority returns the auth authority implementation
func (a *Authorizer) Authority() Authority {
	return a.auth
}

// Transfer changes the auth authority after validation
func (a *Authorizer) Transfer(caller address, newAuthority Authority) error {
	// Ask auth authority to validate the transfer
	return a.auth.Authorize(caller, "transfer_authority", func() error {
		a.auth = newAuthority
		return nil
	})
}

// DoByCurrent executes a privileged action by the auth realm.
func (a *Authorizer) DoByCurrent(title string, action PrivilegedAction, args ...any) error {
	current := runtime.CurrentRealm()
	caller := current.Address()
	return a.auth.Authorize(caller, title, action, args...)
}

// DoByPrevious executes a privileged action by the previous realm.
func (a *Authorizer) DoByPrevious(title string, action PrivilegedAction, args ...any) error {
	previous := runtime.PreviousRealm()
	caller := previous.Address()
	return a.auth.Authorize(caller, title, action, args...)
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

// AddMember adds a new member to the authority
func (a *MemberAuthority) AddMember(caller address, addr address) error {
	return a.Authorize(caller, "add_member", func() error {
		a.members.Add(addr)
		return nil
	})
}

// AddMembers adds a list of members to the authority
func (a *MemberAuthority) AddMembers(caller address, addrs ...address) error {
	return a.Authorize(caller, "add_members", func() error {
		for _, addr := range addrs {
			a.members.Add(addr)
		}
		return nil
	})
}

// RemoveMember removes a member from the authority
func (a *MemberAuthority) RemoveMember(caller address, addr address) error {
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
