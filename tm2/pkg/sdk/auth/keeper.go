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
}

// NewAccountKeeper returns a new AccountKeeper that uses go-amino to
// (binary) encode and decode concrete std.Accounts.
func NewAccountKeeper(
	key store.StoreKey, pk params.ParamsKeeperI, proto func() std.Account,
) AccountKeeper {
	return AccountKeeper{
		key:   key,
		prmk:  pk,
		proto: proto,
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
	stor := ctx.GasStore(ak.key)
	bz := stor.Get(AddressStoreKey(addr))
	if bz == nil {
		return nil
	}
	acc := ak.decodeAccount(bz)
	return acc
}

// GetAllAccounts returns all accounts in the AccountKeeper.
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
	addr := acc.GetAddress()
	stor := ctx.GasStore(ak.key)
	bz, err := amino.MarshalAny(acc)
	if err != nil {
		panic(err)
	}
	stor.Set(AddressStoreKey(addr), bz)
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc std.Account) {
	addr := acc.GetAddress()
	stor := ctx.GasStore(ak.key)
	stor.Delete(AddressStoreKey(addr))
}

// IterateAccounts implements AccountKeeper.
func (ak AccountKeeper) IterateAccounts(ctx sdk.Context, process func(std.Account) (stop bool)) {
	stor := ctx.GasStore(ak.key)
	iter := store.PrefixIterator(stor, []byte(AddressStoreKeyPrefix))
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
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
	var accNumber uint64
	stor := ctx.GasStore(ak.key)
	bz := stor.Get([]byte(GlobalAccountNumberKey))
	if bz == nil {
		accNumber = 0 // start with 0.
	} else {
		err := amino.Unmarshal(bz, &accNumber)
		if err != nil {
			panic(err)
		}
	}

	bz = amino.MustMarshal(accNumber + 1)
	stor.Set([]byte(GlobalAccountNumberKey), bz)

	return accNumber
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
	stor.Set([]byte(GasPriceKey), bz)
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
	bz := stor.Get([]byte(GasPriceKey))
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
