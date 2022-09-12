package vault

import "gno.land/p/grc/grc20"
import "std"
import "time"

type Vault interface {
	Unvault(amount uint)
	Redeem()
	Recover(dest std.Address)
}

func New(token grc20.Token, deposit uint, recover std.Address, lockDuration time.Duration) Vault {
	vault := &impl{
		token:        token,
		recover:      recover,
		lockDuration: lockDuration,

		redeemMinHeight: -1, // will be set in Unvault
		unvaultedAmount: 0,  // will be increased in Unvault, zeroed in Redeem.

	}
	// TODO: owner
	// TODO: deposit
	return vault
}

type impl struct {
	// constructor parameters.
	token        grc20.Token
	recover      std.Address
	lockDuration time.Duration

	// internal parameters.
	owner           std.Address
	redeemMinHeight uint
	unvaultedAmount uint
}

func (*impl) Unvault(amount uint) {
	// assert caller (orig)
	// a
}
func (*impl) Redeem()
func (*impl) Recover(dest std.Address) {}
