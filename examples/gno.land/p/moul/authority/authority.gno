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

// PrivilegedActionHandler is called by contract-based authorities to handle privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// Authority encapsulates ownership and access control.
type Authority struct {
	typeOf          OwnerType
	walletAddr      std.Address             // only when type is OwnerTypeWallet
	contractPath    string                  // only when type is OwnerTypeContract
	contractHandler PrivilegedActionHandler // only when type is OwnerTypeContract
}

// New creates a new Authority instance with the initial owner set to the current caller
// and defaults to a wallet-based admin.
func New() *Authority {
	return &Authority{
		walletAddr: std.PrevRealm().Addr(),
		typeOf:     OwnerTypeWallet,
	}
}

// NewWithAddress creates a new Authority instance with the specified address as the owner.
// The type defaults to a wallet-based admin.
func NewWithAddress(addr std.Address) *Authority {
	return &Authority{
		walletAddr: addr,
		typeOf:     OwnerTypeWallet,
	}
}

// Owner returns the current authority owner.
func (a *Authority) Owner() string {
	if a == nil {
		return ""
	}
	switch a.typeOf {
	case OwnerTypeWallet:
		return a.walletAddr.String()
	case OwnerTypeContract:
		return a.contractPath
	default:
		return ""
	}
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
		caller := std.PrevRealm().Addr()
		if caller != a.walletAddr {
			return errors.New("unauthorized")
		}
		std.Emit(EventSubmit,
			"owner_type", "wallet",
			"owner", a.walletAddr.String(),
			"title", title,
		)
		if err := action(); err != nil {
			return err
		}
		std.Emit(EventExecute,
			"owner_type", "wallet",
			"owner", a.walletAddr.String(),
			"title", title,
		)

	case OwnerTypeContract:
		if a.contractHandler == nil {
			return errors.New("unauthorized: no handler")
		}

		std.Emit(EventExecute,
			"owner_type", "contract",
			"owner", a.contractPath,
			"title", title,
		)

		wrappedAction := func() error {
			caller := std.PrevRealm().PkgPath()
			if caller != a.contractPath {
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
func (a *Authority) TransferToWallet(walletAddr std.Address) error {
	if !walletAddr.IsValid() {
		return errors.New("invalid address")
	}

	prevOwner := a.Owner()
	return a.Do("TransferToWallet", func() error {
		a.walletAddr = walletAddr
		a.contractPath = ""
		a.typeOf = OwnerTypeWallet
		a.contractHandler = nil

		std.Emit("AuthorityTransfer",
			"from", prevOwner,
			"to", walletAddr.String(),
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
		a.walletAddr = ""
		a.contractPath = pkgPath
		a.typeOf = OwnerTypeContract
		a.contractHandler = handler

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
		a.walletAddr = ""
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
