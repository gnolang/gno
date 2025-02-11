// Package authz provides flexible authorization control for privileged actions.
//
// It offers various authorization strategies:
//   - Member-based authorization (single user or teams)
//   - Contract-based authorization (async)
//   - Auto-accept authorization (allow all)
//   - Drop authorization (deny all)
//
// Basic example with member-based authorization:
//
//	import "gno.land/p/moul/authz"
//
//	// Initialize a single-member authority with the contract deployer
//	var auth = authz.New()
//
//	func UpdateConfig(newValue string) error {
//	    return auth.Do("update_config", func() error {
//	        // Only authorized members can execute this
//	        config = newValue
//	        return nil
//	    })
//	}
//
// Example with contract-based authorization (e.g., DAO):
//
//	import (
//	    "gno.land/p/moul/authz"
//	    "gno.land/r/demo/dao"
//	)
//
//	// Create contract authority that delegates to DAO
//	auth := authz.NewContractAuthority("gno.land/r/demo/dao", dao.HandlePrivilegedAction)
//
//	func UpdateParams(params Params) error {
//	    return auth.Do("update_params", func() error {
//	        // This will only execute after DAO approval
//	        return nil
//	    })
//	}
//
// The package provides several authority implementations:
//   - MemberAuthority: Manages a set of authorized addresses
//   - ContractAuthority: Delegates authorization to another contract
//   - AutoAcceptAuthority: Accepts all actions automatically
//   - DroppedAuthority: Denies all actions
//
// Each authority implements the Authority interface which provides:
//   - Do(): Execute privileged actions if authorized
//   - String(): Human-readable description
//
// The main type is Authority, which encapsulates ownership information and
// provides methods for executing privileged actions and transferring
// control. Actions can be executed directly for wallet-based ownership, or
// through a handler for contract-based ownership.
//
// Example usage with wallet ownership:
//
//	// Create a new authority with the current caller as owner
//	var auth = authz.New()
//
//	// Execute a privileged action
//	err := auth.Do("UpdateConfig", func() error {
//	    // Perform privileged operation here
//	    return nil
//	})
//
// Example of transferring from wallet to DAO contract:
//
//	import "gno.land/r/demo/dao"
//
//	// Initial setup with wallet owner
//	var auth = authz.New()
//
//	// Transfer ownership to the DAO contract using its handler
//	err := auth.TransferToContract("gno.land/r/demo/dao", dao.HandlePrivilegedAction)
//
//	// Now all privileged actions must go through the DAO's governance process
//	err = auth.Do("UpdateParams", func() error {
//	    // This action will only execute if there's an approved proposal
//	    return nil
//	})
package authz

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/addrset"
	"gno.land/p/moul/once"
)

// Authorizer is the main wrapper object that handles authority management
type Authorizer struct {
	current Authority
}

// Authority represents an entity that can authorize privileged actions
type Authority interface {
	// Authorize executes a privileged action if the caller is authorized
	Authorize(title string, action PrivilegedAction) error

	// String returns a human-readable description of the authority
	String() string
}

// PrivilegedAction defines a function that performs a privileged action.
type PrivilegedAction func() error

// PrivilegedActionHandler is called by contract-based authorities to handle
// privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// New creates a new Authorizer with the current realm's address as authority
func New() *Authorizer {
	return &Authorizer{
		current: NewMemberAuthority(std.PrevRealm().Addr()),
	}
}

// NewWithAuthority creates a new Authorizer with a specific authority
func NewWithAuthority(authority Authority) *Authorizer {
	return &Authorizer{
		current: authority,
	}
}

// Current returns the current authority implementation
func (a *Authorizer) Current() Authority {
	return a.current
}

// Transfer changes the current authority after validation
func (a *Authorizer) Transfer(newAuthority Authority) error {
	// Ask current authority to validate the transfer
	return a.current.Authorize("transfer_authority", func() error {
		a.current = newAuthority
		return nil
	})
}

// Do executes a privileged action through the current authority
func (a *Authorizer) Do(title string, action PrivilegedAction) error {
	return a.current.Authorize(title, action)
}

// String returns a string representation of the current authority
func (a *Authorizer) String() string {
	return a.current.String()
}

// MemberAuthority is the default implementation using addrset for member management
type MemberAuthority struct {
	members addrset.Set
}

func NewMemberAuthority(members ...std.Address) *MemberAuthority {
	auth := &MemberAuthority{}
	for _, addr := range members {
		auth.members.Add(addr)
	}
	return auth
}

func (a *MemberAuthority) Authorize(title string, action PrivilegedAction) error {
	caller := std.PrevRealm().Addr()
	if !a.members.Has(caller) {
		return errors.New("unauthorized")
	}

	if err := action(); err != nil {
		return err
	}
	return nil
}

func (a *MemberAuthority) String() string {
	return ufmt.Sprintf("member_authority[size=%d]", a.members.Size())
}

// AddMember adds a new member to the authority
func (a *MemberAuthority) AddMember(addr std.Address) error {
	return a.Authorize("add_member", func() error {
		a.members.Add(addr)
		return nil
	})
}

// RemoveMember removes a member from the authority
func (a *MemberAuthority) RemoveMember(addr std.Address) error {
	return a.Authorize("remove_member", func() error {
		a.members.Remove(addr)
		return nil
	})
}

// ListMembers returns the list of all member addresses
// This is a read-only operation that can be called by anyone
func (a *MemberAuthority) ListMembers() []std.Address {
	var members []std.Address
	a.members.IterateByOffset(0, a.members.Size(), func(addr std.Address) bool {
		members = append(members, addr)
		return false // continue iteration
	})
	return members
}

// Has checks if an address is a member
// This is a read-only operation that can be called by anyone
func (a *MemberAuthority) Has(addr std.Address) bool {
	return a.members.Has(addr)
}

// ContractAuthority implements async contract-based authority
type ContractAuthority struct {
	contractPath    string
	contractAddr    std.Address
	contractHandler PrivilegedActionHandler
	proposer        Authority // controls who can create proposals
	executionOnce   once.Once
}

func NewContractAuthority(path string, handler PrivilegedActionHandler) *ContractAuthority {
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    std.DerivePkgAddr(path),
		contractHandler: handler,
		proposer:        NewAutoAcceptAuthority(), // default: anyone can propose
	}
}

func NewRestrictedContractAuthority(path string, handler PrivilegedActionHandler, proposer Authority) *ContractAuthority {
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    std.DerivePkgAddr(path),
		contractHandler: handler,
		proposer:        proposer,
	}
}

func (a *ContractAuthority) Authorize(title string, action PrivilegedAction) error {
	if a.contractHandler == nil {
		return errors.New("contract handler is not set")
	}

	// Wrap the action to ensure it can only be executed by the contract
	wrappedAction := func() error {
		caller := std.PrevRealm().Addr()
		if caller != a.contractAddr {
			return errors.New("action can only be executed by the contract")
		}
		return a.executionOnce.DoErr(func() error {
			return action()
		})
	}

	// Use the proposer authority to control who can create proposals
	return a.proposer.Authorize(title+"_proposal", func() error {
		if err := a.contractHandler(title, wrappedAction); err != nil {
			return err
		}
		return nil
	})
}

func (a *ContractAuthority) String() string {
	return ufmt.Sprintf("contract_authority[contract=%s]", a.contractPath)
}

// AutoAcceptAuthority implements an authority that accepts all actions
// AutoAcceptAuthority is a simple authority that automatically accepts all actions.
// It can be used as a proposer authority to allow anyone to create proposals.
type AutoAcceptAuthority struct{}

func NewAutoAcceptAuthority() *AutoAcceptAuthority {
	return &AutoAcceptAuthority{}
}

func (a *AutoAcceptAuthority) Authorize(title string, action PrivilegedAction) error {
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

func (a *droppedAuthority) Authorize(title string, action PrivilegedAction) error {
	return errors.New("dropped authority: all actions are denied")
}

func (a *droppedAuthority) String() string {
	return "dropped_authority"
}
