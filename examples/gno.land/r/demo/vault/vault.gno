package vault

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
)

type Vault interface {
	Deposit(amount uint, recovery std.Address, lockDuration uint) error
	Unvault(amount uint) error
	Recover(dest std.Address) error
	Redeem() error
}

func New(adminToken *grc20.Banker) Vault {
	return &impl{
		adminToken: adminToken,
		users:      avl.Tree{},
	}
}

type impl struct {
	adminToken *grc20.Banker
	users      avl.Tree // std.Address -> userVault
}

type userVault struct {
	// constructor parameters.
	recover      std.Address
	lockDuration uint

	// internal parameters.
	owner           std.Address
	redeemMinHeight int64
	unvaultedAmount uint
}

func (v *impl) Deposit(amount uint, recover std.Address, lockDuration uint) error {
	caller := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()

	uv := userVault{
		lockDuration:    lockDuration,
		redeemMinHeight: 0, // will be set in Unvault.
		unvaultedAmount: 0, // will be increased in Unvault, zeroed in Redeem.
		owner:           caller,
	}

	// deposit.
	err := v.adminToken.Transfer(caller, pkgAddr, uint64(amount))
	if err != nil {
		return err
	}
	v.users.Set(caller.String(), &uv)

	return nil
}

func (v *impl) Unvault(amount uint) error {
	caller := std.GetOrigCaller()
	uv, err := v.getUserVault(caller)
	if err != nil {
		return err
	}

	balance := v.adminToken.BalanceOf(caller)
	if balance < uint64(amount) {
		return grc20.ErrInsufficientBalance
	}

	println("AAA1", std.GetHeight(), uv.redeemMinHeight, uv.lockDuration)
	uv.redeemMinHeight = std.GetHeight() + int64(uv.lockDuration)
	uv.unvaultedAmount += amount
	v.users.Set(caller.String(), uv)
	println("AAA2", std.GetHeight(), uv.redeemMinHeight, uv.lockDuration)
	return nil
}

func (v *impl) Redeem() error {
	pkgAddr := std.GetOrigPkgAddr()
	caller := std.GetOrigCaller()
	uv, err := v.getUserVault(caller)
	if err != nil {
		return err
	}

	println("AAA3", std.GetHeight(), uv.redeemMinHeight, uv.lockDuration)
	if std.GetHeight() < uv.redeemMinHeight {
		return ErrTooEarlyToRedeem
	}
	// TODO: check balance. (should be optional, but let's be sure).
	// TODO: check height.

	// transfer token.
	err = v.adminToken.Transfer(pkgAddr, caller, uint64(uv.unvaultedAmount))
	if err != nil {
		return err
	}

	uv.unvaultedAmount = 0
	// TODO: if balance == 0 -> destroy?
	return nil
}

func (v *impl) Recover(dest std.Address) error {
	// TODO: assert caller (recovery).
	// TODO: trasfertToken.
	// TODO: destroy?
	return nil
}

func (v *impl) getUserVault(address std.Address) (*userVault, error) {
	uvI, exists := v.users.Get(address.String())
	if !exists {
		return nil, ErrNoSuchVault
	}
	uv := uvI.(*userVault)
	return uv, nil
}
