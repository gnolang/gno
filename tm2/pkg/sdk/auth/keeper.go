package auth

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Concrete implementation of AccountKeeper.
type AccountKeeper struct {
	// The (unexposed) key used to access the store from the Context.
	key store.StoreKey

	// store module parameters
	prmk params.ParamsKeeperI

	// The prototypical Account constructor.
	proto func() std.Account

	// The prototypical SessionAccount constructor.
	sessionProto func() std.Account
}

// NewAccountKeeper returns a new AccountKeeper that uses go-amino to
// (binary) encode and decode concrete std.Accounts.
func NewAccountKeeper(
	key store.StoreKey, pk params.ParamsKeeperI,
	proto func() std.Account,
	sessionProto func() std.Account,
) AccountKeeper {
	return AccountKeeper{
		key:          key,
		prmk:         pk,
		proto:        proto,
		sessionProto: sessionProto,
	}
}

// NewAccountWithAddress implements AccountKeeper.
func (ak AccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) std.Account {
	acc := ak.proto()
	// acc.SetSequence(0) // start with 0.
	err := acc.SetAddress(addr)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	err = acc.SetAccountNumber(ak.GetNextAccountNumber(ctx))
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	return acc
}

// Logger returns a module-specific logger.
func (ak AccountKeeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", ModuleName)
}

// GetAccount returns a specific account in the AccountKeeper.
func (ak AccountKeeper) GetAccount(ctx sdk.Context, addr crypto.Address) std.Account {
	gctx := ctx.GasContext()
	stor := ctx.Store(ak.key)
	bz := stor.Get(gctx, AddressStoreKey(addr))
	if bz == nil {
		return nil
	}
	acc := ak.decodeAccount(bz)
	return acc
}

// GetAllAccounts returns all regular accounts (excludes session accounts).
// Session accounts are stored under the same "/a/" prefix but are filtered
// out by IterateAccounts via key length. Use IterateSessions to access sessions.
func (ak AccountKeeper) GetAllAccounts(ctx sdk.Context) []std.Account {
	accounts := []std.Account{}
	appendAccount := func(acc std.Account) (stop bool) {
		accounts = append(accounts, acc)
		return false
	}
	ak.IterateAccounts(ctx, appendAccount)
	return accounts
}

// SetAccount implements AccountKeeper.
func (ak AccountKeeper) SetAccount(ctx sdk.Context, acc std.Account) {
	gctx := ctx.GasContext()
	addr := acc.GetAddress()
	stor := ctx.Store(ak.key)
	bz, err := amino.MarshalAny(acc)
	if err != nil {
		panic(err)
	}
	stor.Set(gctx, AddressStoreKey(addr), bz)
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc std.Account) {
	gctx := ctx.GasContext()
	addr := acc.GetAddress()
	stor := ctx.Store(ak.key)
	stor.Delete(gctx, AddressStoreKey(addr))
}

// IterateAccounts implements AccountKeeper.
// It iterates over regular accounts only — session accounts (which are
// also stored under the "/a/" prefix at /a/<master>/s/<session>) are
// skipped by checking key length. Regular account keys are exactly
// AccountStoreKeyLen bytes; session sub-keys are longer.
func (ak AccountKeeper) IterateAccounts(ctx sdk.Context, process func(std.Account) (stop bool)) {
	stor := ctx.Store(ak.key)
	iter := store.PrefixIterator(ctx.GasContext(), stor, []byte(AddressStoreKeyPrefix))
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		// Skip session sub-keys. Session accounts are stored at
		// /a/<master>/s/<session> (longer than AccountStoreKeyLen).
		// Regular accounts are at /a/<addr> (exactly AccountStoreKeyLen).
		if len(iter.Key()) != AccountStoreKeyLen {
			iter.Next()
			continue
		}
		val := iter.Value()
		acc := ak.decodeAccount(val)
		if process(acc) {
			return
		}
		iter.Next()
	}
}

// GetPubKey Returns the PubKey of the account at address
func (ak AccountKeeper) GetPubKey(ctx sdk.Context, addr crypto.Address) (crypto.PubKey, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", addr))
	}
	return acc.GetPubKey(), nil
}

// GetSequence Returns the Sequence of the account at address
func (ak AccountKeeper) GetSequence(ctx sdk.Context, addr crypto.Address) (uint64, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", addr))
	}
	return acc.GetSequence(), nil
}

// GetNextAccountNumber Returns and increments the global account number counter
func (ak AccountKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	gctx := ctx.GasContext()
	var accNumber uint64
	stor := ctx.Store(ak.key)
	bz := stor.Get(gctx, []byte(GlobalAccountNumberKey))
	if bz == nil {
		accNumber = 0 // start with 0.
	} else {
		err := amino.Unmarshal(bz, &accNumber)
		if err != nil {
			panic(err)
		}
	}

	bz = amino.MustMarshal(accNumber + 1)
	stor.Set(gctx, []byte(GlobalAccountNumberKey), bz)

	return accNumber
}

// GetSessionAccount returns a session account stored at /a/<master>/s/<session>.
func (ak AccountKeeper) GetSessionAccount(ctx sdk.Context, master, session crypto.Address) std.Account {
	gctx := ctx.GasContext()
	stor := ctx.Store(ak.key)
	bz := stor.Get(gctx, SessionStoreKey(master, session))
	if bz == nil {
		return nil
	}
	var acc std.Account
	err := amino.UnmarshalAny(bz, &acc)
	if err != nil {
		panic(err)
	}
	return acc
}

// SetSessionAccount stores a session account at /a/<master>/s/<session>.
func (ak AccountKeeper) SetSessionAccount(ctx sdk.Context, master crypto.Address, acc std.Account) {
	addr := acc.GetAddress()
	gctx := ctx.GasContext()
	stor := ctx.Store(ak.key)
	bz, err := amino.MarshalAny(acc)
	if err != nil {
		panic(err)
	}
	stor.Set(gctx, SessionStoreKey(master, addr), bz)
}

// RemoveSessionAccount deletes a session account.
func (ak AccountKeeper) RemoveSessionAccount(ctx sdk.Context, master, session crypto.Address) {
	gctx := ctx.GasContext()
	stor := ctx.Store(ak.key)
	stor.Delete(gctx, SessionStoreKey(master, session))
}

// RemoveAllSessions deletes all session accounts for a master via prefix delete.
func (ak AccountKeeper) RemoveAllSessions(ctx sdk.Context, master crypto.Address) {
	gctx := ctx.GasContext()
	stor := ctx.Store(ak.key)
	prefix := SessionPrefixKey(master)
	iter := store.PrefixIterator(gctx, stor, prefix)
	defer iter.Close()
	keys := [][]byte{}
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}
	for _, key := range keys {
		stor.Delete(gctx, key)
	}
}

// IterateSessions iterates over all sessions of a master account.
func (ak AccountKeeper) IterateSessions(ctx sdk.Context, master crypto.Address, cb func(std.Account) bool) {
	gctx := ctx.GasContext()
	stor := ctx.Store(ak.key)
	iter := store.PrefixIterator(gctx, stor, SessionPrefixKey(master))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var acc std.Account
		err := amino.UnmarshalAny(iter.Value(), &acc)
		if err != nil {
			panic(err)
		}
		if cb(acc) {
			break
		}
	}
}

// NewSessionAccount creates a new session account using the session prototype.
func (ak AccountKeeper) NewSessionAccount(ctx sdk.Context, master crypto.Address, pubKey crypto.PubKey) std.Account {
	acc := ak.sessionProto()
	if err := acc.SetAddress(pubKey.Address()); err != nil {
		panic(err)
	}
	if err := acc.SetPubKey(pubKey); err != nil {
		panic(err)
	}
	if err := acc.SetAccountNumber(ak.GetNextAccountNumber(ctx)); err != nil {
		panic(err)
	}
	da := acc.(std.DelegatedAccount)
	if err := da.SetMasterAddress(master); err != nil {
		panic(err)
	}
	return acc
}

// -----------------------------------------------------------------------------
// Misc.
func (ak AccountKeeper) decodeAccount(bz []byte) (acc std.Account) {
	err := amino.Unmarshal(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}

type GasPriceContextKey struct{}

type GasPriceKeeper struct {
	key store.StoreKey
}

// GasPriceKeeper
// The GasPriceKeeper stores the history of gas prices and calculates
// new gas price with formula parameters
func NewGasPriceKeeper(key store.StoreKey) GasPriceKeeper {
	return GasPriceKeeper{
		key: key,
	}
}

// SetGasPrice is called in InitChainer to store initial gas price set in the genesis
func (gk GasPriceKeeper) SetGasPrice(ctx sdk.Context, gp std.GasPrice) {
	if (gp == std.GasPrice{}) {
		return
	}
	stor := ctx.Store(gk.key)
	bz, err := amino.Marshal(gp)
	if err != nil {
		panic(err)
	}
	stor.Set(ctx.GasContext(), []byte(GasPriceKey), bz)
}

// We store the history. If the formula changes, we can replay blocks
// and apply the formula to a specific block range. The new gas price is
// calculated in EndBlock().
func (gk GasPriceKeeper) UpdateGasPrice(ctx sdk.Context) {
	params := ctx.Value(AuthParamsContextKey{}).(Params)
	gasUsed := ctx.BlockGasMeter().GasConsumed()

	// Only update gas price if gas was consumed to avoid changing AppHash
	// on empty blocks.
	if gasUsed <= 0 {
		return
	}

	maxBlockGas := ctx.ConsensusParams().Block.MaxGas
	lgp := gk.LastGasPrice(ctx)
	newGasPrice := gk.calcBlockGasPrice(lgp, gasUsed, maxBlockGas, params)
	gk.SetGasPrice(ctx, newGasPrice)
	logTelemetry(newGasPrice,
		attribute.KeyValue{
			Key:   "func",
			Value: attribute.StringValue("UpdateGasPrice"),
		})
}

// calcBlockGasPrice calculates the minGasPrice for the txs to be included in the next block.
// newGasPrice = lastPrice + lastPrice*(gasUsed-TargetBlockGas)/TargetBlockGas/GasCompressor)
//
// The math formula is an abstraction of a simple solution for the underlying problem we're trying to solve.
// 1. What do we do if the gas used is less than the target gas in a block?
// 2. How do we bring the gas used back to the target level, if gas used is more than the target?
// We simplify the solution with a one-line formula to explain the idea. However, in reality, we need to treat
// two scenarios differently. For example, in the first case, we need to increase the gas by at least 1 unit,
// instead of round down for the integer divisions, and in the second case, we should set a floor
// as the target gas price. This is just a starting point. Down the line, the solution might not be even
// representable by one simple formula
func (gk GasPriceKeeper) calcBlockGasPrice(lastGasPrice std.GasPrice, gasUsed int64, maxGas int64, params Params) std.GasPrice {
	// If no block gas price is set, there is no need to change the last gas price.
	if lastGasPrice.Price.Amount == 0 {
		return lastGasPrice
	}

	// This is also a configuration to indicate that there is no need to change the last gas price.
	if params.TargetGasRatio == 0 {
		return lastGasPrice
	}
	// if no gas used, no need to change the lastPrice
	if gasUsed == 0 {
		return lastGasPrice
	}
	var (
		num   = new(big.Int)
		denom = new(big.Int)
	)

	// targetGas = maxGax*TargetGasRatio/100

	num.Mul(big.NewInt(maxGas), big.NewInt(params.TargetGasRatio))
	num.Div(num, big.NewInt(int64(100)))
	targetGasInt := new(big.Int).Set(num)

	// if used gas is right on target, no need to change
	gasUsedInt := big.NewInt(gasUsed)
	if targetGasInt.Cmp(gasUsedInt) == 0 {
		return lastGasPrice
	}

	c := params.GasPricesChangeCompressor
	lastPriceInt := big.NewInt(lastGasPrice.Price.Amount)

	bigOne := big.NewInt(1)
	if gasUsedInt.Cmp(targetGasInt) == 1 { // gas used is more than the target
		// increase gas price
		num = num.Sub(gasUsedInt, targetGasInt)
		num.Mul(num, lastPriceInt)
		num.Div(num, targetGasInt)
		num.Div(num, denom.SetInt64(c))
		// increase at least 1
		diff := maxBig(num, bigOne)
		num.Add(lastPriceInt, diff)
		// XXX should we cap it with a max gas price?
	} else { // gas used is less than the target
		// decrease gas price down to initial gas price
		initPriceInt := big.NewInt(params.InitialGasPrice.Price.Amount)
		if lastPriceInt.Cmp(initPriceInt) == -1 {
			return params.InitialGasPrice
		}
		num.Sub(targetGasInt, gasUsedInt)
		num.Mul(num, lastPriceInt)
		num.Div(num, targetGasInt)
		num.Div(num, denom.SetInt64(c))

		num.Sub(lastPriceInt, num)
		// gas price should not be less than the initial gas price,
		num = maxBig(num, initPriceInt)
	}

	if !num.IsInt64() {
		panic("The min gas price is out of int64 range")
	}

	lastGasPrice.Price.Amount = num.Int64()
	return lastGasPrice
}

// max returns the larger of x or y.
func maxBig(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}

// It returns the gas price for the last block.
func (gk GasPriceKeeper) LastGasPrice(ctx sdk.Context) std.GasPrice {
	stor := ctx.Store(gk.key)
	// nil GasContext: LastGasPrice is infrastructure/control-plane —
	// it reads the last block's gas price to feed the anteHandler
	// wrapper, the vm/qgasprice query handler, or the EndBlocker.
	// All three callers run under a context whose gas meter is
	// either the default InfiniteGasMeter (queries, EndBlock) or
	// about to be replaced by SetGasMeter inside the auth
	// anteHandler (wrapper path). Charging here would be meaningless
	// or lost; pass nil to skip it.
	bz := stor.Get(nil, []byte(GasPriceKey))
	if bz == nil {
		return std.GasPrice{}
	}

	gp := std.GasPrice{}
	err := amino.Unmarshal(bz, &gp)
	if err != nil {
		panic(err)
	}
	logTelemetry(gp,
		attribute.KeyValue{
			Key:   "func",
			Value: attribute.StringValue("LastGasPrice"),
		})
	return gp
}

func logTelemetry(gp std.GasPrice, kv attribute.KeyValue) {
	if !telemetry.MetricsEnabled() {
		return
	}
	a := attribute.KeyValue{
		Key:   "Coin",
		Value: attribute.StringValue(gp.Price.String()),
	}
	attrs := []attribute.KeyValue{a, kv}
	metrics.BlockGasPriceAmount.Record(
		context.Background(),
		gp.Gas,
		metric.WithAttributes(attrs...),
	)
}
