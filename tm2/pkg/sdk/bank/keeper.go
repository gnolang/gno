package bank

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// bank.Keeper defines a module interface that facilitates the transfer of
// coins between accounts without the possibility of creating coins.
type BankKeeperI interface {
	ViewKeeperI

	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) error
	SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error

	SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error)
	AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error)
	SetCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error

	InitGenesis(ctx sdk.Context, data GenesisState)
	GetParams(ctx sdk.Context) Params
}

var _ BankKeeperI = &BankKeeper{}

// BankKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the BankKeeperI interface.
type BankKeeper struct {
	ViewKeeper

	acck auth.AccountKeeper
	// The keeper used to store parameters
	prmk params.ParamsKeeperI
}

// NewBankKeeper returns a new BankKeeper.
func NewBankKeeper(acck auth.AccountKeeper, pk params.ParamsKeeperI) BankKeeper {
	return BankKeeper{
		ViewKeeper: NewViewKeeper(acck),
		acck:       acck,
		prmk:       pk,
	}
}

// This is a convenience function for manually setting the restricted denoms.
// Useful for testing and initchain setup.
func (bank BankKeeper) SetRestrictedDenoms(ctx sdk.Context, restrictedDenoms []string) {
	bank.prmk.SetStrings(ctx, "p:restricted_denoms", restrictedDenoms)
}

func (bank BankKeeper) RestrictedDenoms(ctx sdk.Context) []string {
	params := bank.GetParams(ctx)
	return params.RestrictedDenoms
}

type stringSet map[string]struct{}

func toSet(str []string) stringSet {
	ss := stringSet{}

	for _, key := range str {
		ss[key] = struct{}{}
	}
	return ss
}

// InputOutputCoins handles a list of inputs and outputs
func (bank BankKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) error {
	// Safety check ensuring that when sending coins the bank must maintain the
	// Check supply invariant and validity of Coins.
	if err := ValidateInputsOutputs(inputs, outputs); err != nil {
		return err
	}

	for _, in := range inputs {
		if !bank.canSendCoins(ctx, in.Address, in.Coins) {
			return std.RestrictedTransferError{}
		}
		// Per-input session spend check: each Input.Address is a tx
		// signer (per MsgMultiSend.GetSigners), so if any input belongs
		// to a session's master, the input amount counts against that
		// session's SpendLimit. No-op for non-session signers.
		if err := auth.CheckAndDeductSessionSpend(ctx, bank.acck, in.Address, in.Coins); err != nil {
			return err
		}
		_, err := bank.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err
		}

		/*
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute(types.AttributeKeySender, in.Address.String()),
				),
			)
		*/
	}

	for _, out := range outputs {
		_, err := bank.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err
		}

		/*
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeTransfer,
					sdk.NewAttribute(types.AttributeKeyRecipient, out.Address.String()),
				),
			)
		*/
	}

	return nil
}

// canSendCoins returns true if the coins can be sent without violating any restriction.
func (bank BankKeeper) canSendCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool {
	rds := bank.RestrictedDenoms(ctx)
	if len(rds) == 0 {
		// No restrictions.
		return true
	}
	if amt.ContainOneOfDenom(toSet(rds)) {
		acc := bank.acck.GetAccount(ctx, addr)
		accr, ok := acc.(std.AccountUnrestricter)
		if ok && accr.IsTokenLockWhitelisted() {
			return true
		}
		return false
	}
	return true
}

// SendCoins moves coins from one account to another, restrction could be applied
func (bank BankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	// If amt is zero do nothing.
	if amt.IsZero() {
		return nil
	}

	// read restricted boolean value from param.IsRestrictedTransfer()
	// canSendCoins is true until they have agreed to the waiver
	if !bank.canSendCoins(ctx, fromAddr, amt) {
		return std.RestrictedTransferError{}
	}

	// If the tx is session-signed and fromAddr is the session's master,
	// deduct from the session's SpendLimit. No-op otherwise.
	// SendCoinsUnrestricted deliberately bypasses this (gas collection,
	// storage deposit refunds).
	if err := auth.CheckAndDeductSessionSpend(ctx, bank.acck, fromAddr, amt); err != nil {
		return err
	}

	return bank.sendCoins(ctx, fromAddr, toAddr, amt)
}

// SendCoinsUnrestricted is used for paying gas.
// Unvested coins cannot be used.
func (bank BankKeeper) SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	_, err := bank.subtractCoinsUnrestricted(ctx, fromAddr, amt)
	if err != nil {
		return err
	}
	_, err = bank.AddCoins(ctx, toAddr, amt)
	return err
}

func (bank BankKeeper) sendCoins(
	ctx sdk.Context,
	fromAddr crypto.Address,
	toAddr crypto.Address,
	amt std.Coins,
) error {
	_, err := bank.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	_, err = bank.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	/*
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, toAddr.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
			),
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
			),
		})
	*/

	return nil
}

// upgradeVestingAccount replaces a fully-vested VestingAccount with a plain
// BaseAccount. Returns the replacement account, or the original if not ready.
func (bank BankKeeper) upgradeVestingAccount(ctx sdk.Context, acc std.Account) std.Account {
	va, ok := acc.(std.VestingAccount)
	if !ok {
		return acc
	}
	if !va.GetVestingCoins(ctx.BlockTime()).IsZero() {
		return acc
	}
	baseAcc := &std.BaseAccount{
		Address:       va.GetAddress(),
		Coins:         va.GetCoins(),
		PubKey:        va.GetPubKey(),
		AccountNumber: va.GetAccountNumber(),
		Sequence:      va.GetSequence(),
	}
	bank.acck.SetAccount(ctx, baseAcc)
	return baseAcc
}

// SubtractCoins subtracts amt from the coins at the addr.
//
// Enforces vesting: if the account is a VestingAccount, the amount must
// not exceed the spendable (unlocked) coins at the current block time.
func (bank BankKeeper) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	if !amt.IsValid() {
		return nil, std.ErrInvalidCoins(amt.String())
	}

	oldCoins := std.NewCoins()
	acc := bank.acck.GetAccount(ctx, addr)
	if acc != nil {
		oldCoins = acc.GetCoins()
	}

	// Vesting enforcement: the amount subtracted must be spendable.
	// Once the vesting schedule completes, the account is upgraded to a
	// plain BaseAccount so future transfers skip vesting checks entirely.
	acc = bank.upgradeVestingAccount(ctx, acc)
	if va, ok := acc.(std.VestingAccount); ok {
		spendable := std.SpendableCoins(va, ctx.BlockTime())
		if !spendable.IsAllGTE(amt) {
			return nil, std.ErrVestingLockedCoins(fmt.Sprintf(
				"insufficient spendable coins; %s < %s (locked=%s)",
				spendable, amt, va.LockedCoins(ctx.BlockTime()),
			))
		}
	}

	newCoins := oldCoins.SubUnsafe(amt)
	if !newCoins.IsValid() {
		err := std.ErrInsufficientCoins(
			fmt.Sprintf("insufficient account funds; %s < %s", oldCoins, amt),
		)
		return nil, err
	}
	err := bank.SetCoins(ctx, addr, newCoins)

	return newCoins, err
}

// subtractCoinsUnrestricted performs raw coin subtraction without vesting or
// session-spend enforcement. Used for gas payments, storage deposit refunds,
// and other system-level transfers.
//
// Still upgrades fully-vested accounts to BaseAccount when the schedule ends.
func (bank BankKeeper) subtractCoinsUnrestricted(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	if !amt.IsValid() {
		return nil, std.ErrInvalidCoins(amt.String())
	}

	oldCoins := std.NewCoins()
	acc := bank.acck.GetAccount(ctx, addr)
	if acc != nil {
		oldCoins = acc.GetCoins()
	}

	// Upgrade fully-vested accounts even on unrestricted transfers.
	bank.upgradeVestingAccount(ctx, acc)

	newCoins := oldCoins.SubUnsafe(amt)
	if !newCoins.IsValid() {
		err := std.ErrInsufficientCoins(
			fmt.Sprintf("insufficient account funds; %s < %s", oldCoins, amt),
		)
		return nil, err
	}
	err := bank.SetCoins(ctx, addr, newCoins)

	return newCoins, err
}

// AddCoins adds amt to the coins at the addr.
func (bank BankKeeper) AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	if !amt.IsValid() {
		return nil, std.ErrInvalidCoins(amt.String())
	}

	oldCoins := bank.GetCoins(ctx, addr)
	newCoins := oldCoins.Add(amt)

	if !newCoins.IsValid() {
		return amt, std.ErrInsufficientCoins(
			fmt.Sprintf("insufficient account funds; %s < %s", oldCoins, amt),
		)
	}

	err := bank.SetCoins(ctx, addr, newCoins)
	return newCoins, err
}

// SetCoins sets the coins at the addr.
func (bank BankKeeper) SetCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}

	acc := bank.acck.GetAccount(ctx, addr)
	if acc == nil {
		acc = bank.acck.NewAccountWithAddress(ctx, addr)
	}

	err := acc.SetCoins(amt)
	if err != nil {
		panic(err)
	}

	bank.acck.SetAccount(ctx, acc)
	return nil
}

// ----------------------------------------
// ViewKeeper

// ViewKeeperI defines a module interface that facilitates read only access to
// account balances.
type ViewKeeperI interface {
	GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins
	HasCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool
}

var _ ViewKeeperI = ViewKeeper{}

// ViewKeeper implements a read only keeper implementation of ViewKeeperI.
type ViewKeeper struct {
	acck auth.AccountKeeper
}

// NewViewKeeper returns a new ViewKeeper.
func NewViewKeeper(acck auth.AccountKeeper) ViewKeeper {
	return ViewKeeper{acck: acck}
}

// Logger returns a module-specific logger.
func (view ViewKeeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}

// GetCoins returns the coins at the addr.
func (view ViewKeeper) GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins {
	acc := view.acck.GetAccount(ctx, addr)
	if acc == nil {
		return std.NewCoins()
	}
	return acc.GetCoins()
}

// HasCoins returns whether or not an account has at least amt coins.
func (view ViewKeeper) HasCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool {
	return view.GetCoins(ctx, addr).IsAllGTE(amt)
}
