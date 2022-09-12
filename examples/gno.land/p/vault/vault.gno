package vault

import (
	"gno.land/p/grc/grc20"
	"std"
	"time"
)

type Vault interface {
	Unvault(amount uint)
	Redeem()
	Recover(dest std.Address)
}

func New(token grc20.IGRC20, deposit uint, recover std.Address, lockDuration time.Duration) Vault {
	vault := &impl{}
	/*
		token:        token,
		recover:      recover,
		lockDuration: lockDuration,

		redeemMinHeight: -1, // will be set in Unvault.
		unvaultedAmount: 0,  // will be increased in Unvault, zeroed in Redeem.
		owner:           std.GetOrigCaller(),
	*/
	// TODO: owner
	// TODO: deposit
	return vault
}

type impl struct {
	// constructor parameters.
	token        grc20.IGRC20
	recover      std.Address
	lockDuration time.Duration

	// internal parameters.
	owner           std.Address
	redeemMinHeight uint
	unvaultedAmount uint
}

func (v *impl) Unvault(amount uint) {
	// TODO: assert caller (orig).
	// TODO: check balance.
	// TODO: set redeemMinHeight = now + durationToHeight(lockDuration() + 1.
	v.unvaultedAmount += amount
}

func (v *impl) Redeem() {
	// TODO: assert caller.
	// TODO: check balance. (should be optional, but let's be sure).
	// TODO: check height.
	// TODO: transferToken.
	// TODO: if balance == 0 -> destroy?
}

func (v *impl) Recover(dest std.Address) {
	// TODO: assert caller (recovery).
	// TODO: trasfertToken.
	// TODO: destroy?
}
