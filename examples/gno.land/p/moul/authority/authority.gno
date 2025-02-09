// Package authority provides a flexible ownership and access control system for
// Gno contracts.
//
// It supports both wallet-based and contract-based ownership models, allowing
// for secure management of privileged operations. The package enables ownership
// transfers between wallets and contracts, and includes support for dropping
// ownership entirely.
//
// The main type is Authority, which encapsulates ownership information and
// provides methods for executing privileged actions and transferring
// control. Actions can be executed directly for wallet-based ownership, or
// through a handler for contract-based ownership.
//
// Example usage with wallet ownership:
//
//	// Create a new authority with the current caller as owner
//	auth := authority.New()
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
//	auth := authority.New()
//
//	// Transfer ownership to the DAO contract using its handler
//	err := auth.TransferToContract("gno.land/r/demo/dao", dao.HandlePrivilegedAction)
//
//	// Now all privileged actions must go through the DAO's governance process
//	err = auth.Do("UpdateParams", func() error {
//	    // This action will only execute if there's an approved proposal
//	    return nil
//	})
package authority

import (
	"errors"
	"std"
)

type OwnerType uint8

const (
	OwnerTypeUnset OwnerType = iota
	OwnerTypeDropped
	OwnerTypeWallet
	OwnerTypeContract
)

const (
	EventSubmit  = "submit"
	EventExecute = "execute"
)

// PrivilegedAction defines a function that performs a privileged action.
type PrivilegedAction func() error

// PrivilegedActionHandler is called by contract-based authorities to handle
// privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// Authority encapsulates ownership and access control.
type Authority struct {
	typeOf          OwnerType
	ownerAddr       std.Address
	contractPath    string                  // only when type is OwnerTypeContract
	contractHandler PrivilegedActionHandler // only when type is OwnerTypeContract
}

// New creates a new Authority instance with the initial owner set to the
// current caller and defaults to a wallet-based admin.
func New() *Authority {
	return &Authority{
		ownerAddr: std.PrevRealm().Addr(),
		typeOf:    OwnerTypeWallet,
	}
}

// NewWithAddress creates a new Authority instance with the specified address as
// the owner. The type defaults to a wallet-based admin.
func NewWithAddress(addr std.Address) *Authority {
	return &Authority{
		ownerAddr: addr,
		typeOf:    OwnerTypeWallet,
	}
}

// NewWithContract creates a new Authority instance with the specified contract
// as the owner.
func NewWithContract(pkgPath string, handler PrivilegedActionHandler) (*Authority, error) {
	if pkgPath == "" {
		return nil, errors.New("invalid package path")
	}
	if handler == nil {
		return nil, errors.New("handler required")
	}

	return &Authority{
		ownerAddr:       std.DerivePkgAddr(pkgPath),
		typeOf:          OwnerTypeContract,
		contractPath:    pkgPath,
		contractHandler: handler,
	}, nil
}

// Owner returns the current authority owner.
func (a *Authority) Owner() string {
	if a == nil {
		return ""
	}
	switch a.typeOf {
	case OwnerTypeWallet:
		return a.ownerAddr.String()
	case OwnerTypeContract:
		return a.contractPath
	default:
		return ""
	}
}

// Do executes a privileged action.
// For wallet-based admins, the action is executed immediately if the caller is
// authorized.
// For contract-based authorities, the action is passed to the handler for
// processing.
func (a *Authority) Do(title string, action PrivilegedAction) error {
	if a == nil {
		return errors.New("unauthorized")
	}

	switch a.typeOf {
	case OwnerTypeWallet:
		caller := std.PrevRealm().Addr()
		if caller != a.ownerAddr {
			return errors.New("unauthorized")
		}
		std.Emit(EventSubmit,
			"owner_type", "wallet",
			"owner", a.ownerAddr.String(),
			"title", title,
		)
		if err := action(); err != nil {
			return err
		}
		std.Emit(EventExecute,
			"owner_type", "wallet",
			"owner", a.ownerAddr.String(),
			"title", title,
		)

	case OwnerTypeContract:
		if a.contractHandler == nil {
			return errors.New("unauthorized: no handler")
		}

		std.Emit(EventSubmit,
			"owner_type", "contract",
			"owner", a.contractPath,
			"title", title,
		)

		wrappedAction := func() error {
			caller := std.PrevRealm().Addr()
			if caller != a.ownerAddr {
				return errors.New("unauthorized: invalid contract caller")
			}
			if ret := action(); ret != nil {
				return ret
			}
			std.Emit(EventExecute,
				"owner_type", "contract",
				"owner", a.contractPath,
				"title", title,
			)
			return nil
		}

		if err := a.contractHandler(title, wrappedAction); err != nil {
			return err
		}

	default:
		return errors.New("unauthorized: invalid owner type")
	}

	return nil
}

// TransferToWallet transfers authority to a new wallet address.
func (a *Authority) TransferToWallet(ownerAddr std.Address) error {
	if !ownerAddr.IsValid() {
		return errors.New("invalid address")
	}

	prevOwner := a.Owner()
	return a.Do("TransferToWallet", func() error {
		a.ownerAddr = ownerAddr
		a.contractPath = ""
		a.typeOf = OwnerTypeWallet
		a.contractHandler = nil

		std.Emit("AuthorityTransfer",
			"from", prevOwner,
			"to", ownerAddr.String(),
			"type", "wallet",
		)
		return nil
	})
}

// TransferToContract transfers authority to a contract address with a handler.
func (a *Authority) TransferToContract(pkgPath string, handler PrivilegedActionHandler) error {
	if pkgPath == "" {
		return errors.New("invalid package path")
	}
	if handler == nil {
		return errors.New("handler required")
	}

	prevOwner := a.Owner()
	return a.Do("TransferToContract", func() error {
		a.ownerAddr = ""
		a.contractPath = pkgPath
		a.typeOf = OwnerTypeContract
		a.contractHandler = handler
		a.ownerAddr = std.DerivePkgAddr(pkgPath)

		std.Emit("AuthorityTransfer",
			"from", prevOwner,
			"to", pkgPath,
			"type", "contract",
		)
		return nil
	})
}

// Drop revokes all authority by setting the owner to an empty address.
func (a *Authority) Drop() error {
	prevOwner := a.Owner()
	return a.Do("Drop", func() error {
		a.ownerAddr = ""
		a.contractPath = ""
		a.typeOf = OwnerTypeDropped
		a.contractHandler = nil

		std.Emit("AuthorityTransfer",
			"from", prevOwner,
			"to", "",
			"type", "dropped",
		)
		return nil
	})
}
