package bank

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
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

	TotalCoin(ctx sdk.Context, denom string) int64

	InitGenesis(ctx sdk.Context, data GenesisState)
	GetParams(ctx sdk.Context) Params
}

var _ BankKeeperI = &BankKeeper{}

// BankKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the BankKeeperI interface.
type BankKeeper struct {
	ViewKeeper

	key  store.StoreKey
	acck auth.AccountKeeper
	// The keeper used to store parameters
	prmk params.ParamsKeeperI
}

// NewBankKeeper returns a new BankKeeper.
func NewBankKeeper(key store.StoreKey, acck auth.AccountKeeper, pk params.ParamsKeeperI) BankKeeper {
	return BankKeeper{
		ViewKeeper: NewViewKeeper(acck),
		key:        key,
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
	// read restricted boolean value from param.IsRestrictedTransfer()
	// canSendCoins is true until they have agreed to the waiver
	if !bank.canSendCoins(ctx, fromAddr, amt) {
		return std.RestrictedTransferError{}
	}

	return bank.sendCoins(ctx, fromAddr, toAddr, amt)
}

// SendCoinsUnrestricted is used for paying gas.
func (bank BankKeeper) SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	return bank.sendCoins(ctx, fromAddr, toAddr, amt)
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

// SubtractCoins subtracts amt from the coins at the addr.
//
// CONTRACT: If the account is a vesting account, the amount has to be spendable.
func (bank BankKeeper) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	if !amt.IsValid() {
		return nil, std.ErrInvalidCoins(amt.String())
	}

	oldCoins := std.NewCoins()
	acc := bank.acck.GetAccount(ctx, addr)
	if acc != nil {
		oldCoins = acc.GetCoins()
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

	// Update supply index: compute delta between old and new balances.
	oldCoins := acc.GetCoins()
	bank.updateSupply(ctx, oldCoins, amt)

	err := acc.SetCoins(amt)
	if err != nil {
		panic(err)
	}

	bank.acck.SetAccount(ctx, acc)
	return nil
}

// updateSupply adjusts the per-denomination supply index based on the delta
// between the old and new coin balances for an account.
// Panics on overflow, as exceeding int64 range indicates a broken invariant.
func (bank BankKeeper) updateSupply(ctx sdk.Context, oldCoins, newCoins std.Coins) {
	// Collect all denoms that appear in either old or new.
	denoms := make(map[string]struct{})
	for _, c := range oldCoins {
		denoms[c.Denom] = struct{}{}
	}
	for _, c := range newCoins {
		denoms[c.Denom] = struct{}{}
	}

	stor := ctx.Store(bank.key)
	for denom := range denoms {
		oldAmt := oldCoins.AmountOf(denom)
		newAmt := newCoins.AmountOf(denom)
		delta, ok := overflow.Sub(newAmt, oldAmt)
		if !ok {
			panic(fmt.Sprintf("supply delta overflow for denom %q: %d - %d", denom, newAmt, oldAmt))
		}
		if delta == 0 {
			continue
		}
		supply := bank.getSupply(stor, denom)
		newSupply, ok := overflow.Add(supply, delta)
		if !ok {
			panic(fmt.Sprintf("total supply overflow for denom %q: %d + %d", denom, supply, delta))
		}
		bank.setSupply(stor, denom, newSupply)
	}
}

// getSupply reads the total supply of a denomination from the store.
func (bank BankKeeper) getSupply(stor store.Store, denom string) int64 {
	bz := stor.Get(SupplyStoreKey(denom))
	if bz == nil {
		return 0
	}
	var supply int64
	err := amino.Unmarshal(bz, &supply)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal supply for denom %q: %w", denom, err))
	}
	return supply
}

// setSupply writes the total supply of a denomination to the store.
// If supply is zero, the key is deleted to avoid storing nil values.
func (bank BankKeeper) setSupply(stor store.Store, denom string, supply int64) {
	key := SupplyStoreKey(denom)
	if supply == 0 {
		stor.Delete(key)
		return
	}
	bz := amino.MustMarshal(supply)
	stor.Set(key, bz)
}

// TotalCoin returns the total supply of a given coin denomination.
// This is an O(1) read from the supply index.
func (bank BankKeeper) TotalCoin(ctx sdk.Context, denom string) int64 {
	stor := ctx.Store(bank.key)
	return bank.getSupply(stor, denom)
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
