package vault

import (
	"std"

	"gno.land/p/grc/grc20"
)

type Vault interface {
	Unvault(amount uint) error
	Redeem() error
	Recover(dest std.Address) error
}

func New(token grc20.IGRC20, deposit uint, recover std.Address, lockDuration uint) Vault {
	vault := &impl{
		token:        token,
		recover:      recover,
		lockDuration: lockDuration,

		redeemMinHeight: 0, // will be set in Unvault.
		unvaultedAmount: 0, // will be increased in Unvault, zeroed in Redeem.
		owner:           std.GetOrigCaller(),
	}

	// deposit.
	pkgAddr := std.GetOrigPkgAddr()
	err := token.Transfer(pkgAddr, uint64(deposit))
	if err != nil {
		panic(err)
	}
	// TODO: deposit
	return vault
}

type impl struct {
	// constructor parameters.
	token        grc20.IGRC20
	recover      std.Address
	lockDuration uint

	// internal parameters.
	owner           std.Address
	redeemMinHeight uint
	unvaultedAmount uint
}

func (v *impl) Unvault(amount uint) error {
	// TODO: assert caller (orig).
	// TODO: check balance.
	// TODO: set redeemMinHeight = now + lockDuration()
	v.unvaultedAmount += amount
	return nil
}

func (v *impl) Redeem() error {
	// TODO: assert caller.
	// TODO: check balance. (should be optional, but let's be sure).
	// TODO: check height.
	// TODO: transferToken.
	// TODO: if balance == 0 -> destroy?
	return nil
}

func (v *impl) Recover(dest std.Address) error {
	// TODO: assert caller (recovery).
	// TODO: trasfertToken.
	// TODO: destroy?
	return nil
}
