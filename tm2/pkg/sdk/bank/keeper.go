package bank

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
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
}

var _ BankKeeperI = BankKeeper{}

// BBankKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the BankKeeper interface.
type BankKeeper struct {
	ViewKeeper

	acck auth.AccountKeeper
	tck  TotalCoinKeeper
}

// NewBankKeeper returns a new BankKeeper.
func NewBankKeeper(acck auth.AccountKeeper, tck TotalCoinKeeper) BankKeeper {
	return BankKeeper{
		ViewKeeper: NewViewKeeper(acck, tck),
		acck:       acck,
		tck:        tck,
	}
}

// InputOutputCoins handles a list of inputs and outputs
func (bank BankKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) error {
	// Safety check ensuring that when sending coins the bank must maintain the
	// Check supply invariant and validity of Coins.
	if err := ValidateInputsOutputs(inputs, outputs); err != nil {
		return err
	}

	for _, in := range inputs {
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

// SendCoins moves coins from one account to another
func (bank BankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
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
	if err != nil {
		return nil, err
	}

	err = bank.tck.decreaseTotalCoin(ctx, amt)
	if err != nil {
		return nil, err
	}

	return newCoins, nil
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
	if err != nil {
		return nil, err
	}

	err = bank.tck.increaseTotalCoin(ctx, amt)
	if err != nil {
		return nil, err
	}

	return newCoins, nil
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
	TotalCoin(ctx sdk.Context, denom string) int64
}

var _ ViewKeeperI = ViewKeeper{}

// ViewKeeper implements a read only keeper implementation of ViewKeeperI.
type ViewKeeper struct {
	acck auth.AccountKeeper
	tck  TotalCoinKeeper
}

// NewViewKeeper returns a new ViewKeeper.
func NewViewKeeper(acck auth.AccountKeeper, tck TotalCoinKeeper) ViewKeeper {
	return ViewKeeper{acck: acck, tck: tck}
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

// TotalCoin returns the total coin for a given denomination.
func (view ViewKeeper) TotalCoin(ctx sdk.Context, denom string) int64 {
	return view.tck.TotalCoin(ctx, denom)
}

// TotalCoinKeeper manages the total amount of coins for various denominations.
type TotalCoinKeeper struct {
	key store.StoreKey
}

// NewTotalCoinKeeper returns a new TotalCoinKeeper.
func NewTotalCoinKeeper(key store.StoreKey) TotalCoinKeeper {
	return TotalCoinKeeper{
		key: key,
	}
}

// TotalCoin returns the total coin for a given denomination.
func (tck TotalCoinKeeper) TotalCoin(ctx sdk.Context, denom string) int64 {
	stor := ctx.Store(tck.key)
	bz := stor.Get(TotalCoinStoreKey(denom))
	if bz == nil {
		return 0
	}

	totalCoin := tck.decodeTotalCoin(bz)

	return totalCoin.Amount
}

// increaseTotalCoin increases the total coin amounts for the specified denominations.
func (tck TotalCoinKeeper) increaseTotalCoin(ctx sdk.Context, coins std.Coins) error {
	stor := ctx.Store(tck.key)

	for _, coin := range coins {
		oldTotalCoin := std.NewCoin(coin.Denom, 0)

		bz := stor.Get(TotalCoinStoreKey(coin.Denom))
		if bz != nil {
			oldTotalCoin = tck.decodeTotalCoin(bz)
		}

		newTotalCoin := oldTotalCoin.Add(coin)
		err := tck.setTotalCoin(ctx, newTotalCoin)
		if err != nil {
			return err
		}
	}

	return nil
}

// decreaseTotalCoin decreases the total coin amounts for the specified denominations.
func (tck TotalCoinKeeper) decreaseTotalCoin(ctx sdk.Context, coins std.Coins) error {
	stor := ctx.Store(tck.key)

	for _, coin := range coins {
		bz := stor.Get(TotalCoinStoreKey(coin.Denom))
		if bz == nil {
			return std.ErrInvalidCoins(fmt.Sprintf("denomination %s not found", coin.Denom))
		}

		oldTotalCoin := tck.decodeTotalCoin(bz)
		newTotalCoin := oldTotalCoin.SubUnsafe(coin)
		if !newTotalCoin.IsValid() {
			err := std.ErrInsufficientCoins(
				fmt.Sprintf("insufficient account funds; %s < %s", oldTotalCoin, coin),
			)
			return err
		}

		err := tck.setTotalCoin(ctx, newTotalCoin)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetTotalCoin sets the total coin amount for a given denomination.
func (tck TotalCoinKeeper) setTotalCoin(ctx sdk.Context, totalCoin std.Coin) error {
	stor := ctx.Store(tck.key)
	bz, err := totalCoin.MarshalAmino()
	if err != nil {
		return err
	}
	stor.Set(TotalCoinStoreKey(totalCoin.Denom), []byte(bz))
	return nil
}

// decodeTotalCoin decodes the total coin from a byte slice.
func (tck TotalCoinKeeper) decodeTotalCoin(bz []byte) std.Coin {
	var coin std.Coin
	err := coin.UnmarshalAmino(string(bz))
	if err != nil {
		panic(err)
	}
	return coin
}
