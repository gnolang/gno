package vm

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store"
)

// vm.VMKeeperI defines a module interface that supports Gno
// smart contracts programming (scripting).
type VMKeeperI interface {
	AddPackage(ctx sdk.Context, creator crypto.Address, pkgPath string, files []NamedFile) error
	Exec(ctx sdk.Context, caller crypto.Address, stmt string) error
}

var _ VMKeeperI = VMKeeper{}

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	key  store.StoreKey
	acck auth.AccountKeeper
	bank bank.BankKeeper
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(key store.StoreKey, acck auth.AccountKeeper, bank bank.BankKeeper) VMKeeper {
	return VMKeeper{
		key:  key,
		acck: acck,
		bank: bank,
	}
}

// AddPackage adds a package with given fileset.
func (vm VMKeeper) AddPackage(ctx sdk.Context, creator crypto.Address, pkgPath string, files []NamedFile) error {
	// Validate arguments.
	if creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	creatorAcc := vm.acck.GetAccount(ctx, creator)
	if creatorAcc == nil {
		return std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", creator))
	}
	if pkgPath == "" {
		return ErrInvalidPkgPath("missing package path")
	}
	// TODO check to ensure that package name doesn't already exist.
	// TODO check to ensure that creator can pay.
	// TODO deduct price from creator.
	// TODO add files to global. (hack)
	// TODO parse and run the files.
	return nil
}

// Exec executes limited forms of gno statements.
func (vm VMKeeper) Exec(ctx sdk.Context, caller crypto.Address, stmt string) error {
	// TODO pay for gas? TODO see context?
	/*
		_, err := vm.SubtractCoins(ctx, caller, amt)
		if err != nil {
			return err
		}
	*/
	return nil
}
