package authority

import (
	"errors"
	"std"
)

type OwnerType uint8

const (
	OwnerTypeUnset OwnerType = iota
	OwnerTypeWallet
	OwnerTypeContract
)

const (
	ActionEvent = "AuthorityAction"
)

// PrivilegedAction defines a function that performs a privileged action.
type PrivilegedAction func() error

// PrivilegedActionHandler is called by contract-based authorities to handle privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// Authority encapsulates an owner and its type.
// It supports both wallet-based ownership (admin) and contract-based ownership.
type Authority struct {
	owner   std.Address
	typeOf  OwnerType
	handler PrivilegedActionHandler
}

// New creates a new Authority instance with the initial owner set to the current caller
// and defaults to a wallet-based admin.
func New() *Authority {
	return &Authority{
		owner:  std.PrevRealm().Addr(),
		typeOf: OwnerTypeWallet,
	}
}

// NewWithAddress creates a new Authority instance with the specified address as the owner.
// The type defaults to a wallet-based admin.
func NewWithAddress(addr std.Address) *Authority {
	return &Authority{
		owner:  addr,
		typeOf: OwnerTypeWallet,
	}
}

// Owner returns the current authority owner.
func (a *Authority) Owner() std.Address {
	if a == nil {
		return ""
	}
	return a.owner
}

// Do executes a privileged action.
// For wallet-based admins, the action is executed immediately if the caller is authorized.
// For contract-based authorities, the action is passed to the handler for processing.
func (a *Authority) Do(title string, action PrivilegedAction) error {
	if a == nil {
		return errors.New("unauthorized")
	}

	switch a.typeOf {
	case OwnerTypeWallet:
		if std.PrevRealm().Addr() != a.owner {
			return errors.New("unauthorized")
		}
		if err := action(); err != nil {
			return err
		}
		std.Emit(ActionEvent,
			"type", "wallet",
			"owner", a.owner.String(),
			"title", title,
		)
		return nil

	case OwnerTypeContract:
		if a.handler == nil {
			return errors.New("unauthorized: no handler")
		}

		wrappedAction := func() error {
			caller := std.PrevRealm().Addr()
			if caller != a.owner {
				panic("unauthorized: invalid contract caller")
			}
			if err := action(); err != nil {
				return err
			}
			std.Emit(ActionEvent,
				"type", "contract",
				"owner", a.owner.String(),
				"title", title,
			)
			return nil
		}

		if err := a.handler(title, wrappedAction); err != nil {
			return err
		}

	default:
		return errors.New("unauthorized: invalid owner type")
	}

	return nil
}

// TransferToWallet transfers authority to a new wallet address.
func (a *Authority) TransferToWallet(walletAddr std.Address) error {
	if !walletAddr.IsValid() {
		return errors.New("invalid address")
	}

	prevOwner := a.owner
	return a.Do("TransferToWallet", func() error {
		a.owner = walletAddr
		a.typeOf = OwnerTypeWallet
		a.handler = nil

		std.Emit("AuthorityTransfer",
			"from", prevOwner.String(),
			"to", walletAddr.String(),
			"type", "wallet",
		)
		return nil
	})
}

// TransferToContract transfers authority to a contract address with a handler.
func (a *Authority) TransferToContract(contractAddr std.Address, handler PrivilegedActionHandler) error {
	if !contractAddr.IsValid() {
		return errors.New("invalid address")
	}
	if handler == nil {
		return errors.New("handler required")
	}

	prevOwner := a.owner
	return a.Do("TransferToContract", func() error {
		a.owner = contractAddr
		a.typeOf = OwnerTypeContract
		a.handler = handler

		std.Emit("AuthorityTransfer",
			"from", prevOwner.String(),
			"to", contractAddr.String(),
			"type", "contract",
		)
		return nil
	})
}

// Drop revokes all authority by setting the owner to an empty address.
func (a *Authority) Drop() error {
	prevOwner := a.owner
	return a.Do("Drop", func() error {
		a.owner = ""
		a.typeOf = OwnerTypeUnset
		a.handler = nil

		std.Emit("AuthorityTransfer",
			"from", prevOwner.String(),
			"to", "",
		)
		return nil
	})
}
