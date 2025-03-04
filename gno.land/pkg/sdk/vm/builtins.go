package vm

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ----------------------------------------
// SDKBanker

type SDKBanker struct {
	vmk *VMKeeper
	ctx sdk.Context
}

func NewSDKBanker(vmk *VMKeeper, ctx sdk.Context) *SDKBanker {
	return &SDKBanker{
		vmk: vmk,
		ctx: ctx,
	}
}

func (bnk *SDKBanker) GetCoins(b32addr crypto.Bech32Address) (dst std.Coins) {
	addr := crypto.MustAddressFromString(string(b32addr))
	coins := bnk.vmk.bank.GetCoins(bnk.ctx, addr)
	return coins
}

func (bnk *SDKBanker) SendCoins(b32from, b32to crypto.Bech32Address, amt std.Coins) {
	from := crypto.MustAddressFromString(string(b32from))
	to := crypto.MustAddressFromString(string(b32to))
	err := bnk.vmk.bank.SendCoins(bnk.ctx, from, to, amt)
	if err != nil {
		panic(err)
	}
}

func (bnk *SDKBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

func (bnk *SDKBanker) IssueCoin(b32addr crypto.Bech32Address, denom string, amount int64) {
	addr := crypto.MustAddressFromString(string(b32addr))
	_, err := bnk.vmk.bank.AddCoins(bnk.ctx, addr, std.Coins{std.Coin{Denom: denom, Amount: amount}})
	if err != nil {
		panic(err)
	}
}

func (bnk *SDKBanker) RemoveCoin(b32addr crypto.Bech32Address, denom string, amount int64) {
	addr := crypto.MustAddressFromString(string(b32addr))
	_, err := bnk.vmk.bank.SubtractCoins(bnk.ctx, addr, std.Coins{std.Coin{Denom: denom, Amount: amount}})
	if err != nil {
		panic(err)
	}
}

// ----------------------------------------
// SDKParams

// This implements ParamsInterface,
// which is available as ExecContext.Params.
// Access to SDKParams gives access to all parameters.
// Users must write code to limit access as appropriate.

type SDKParams struct {
	pmk params.ParamsKeeper
	ctx sdk.Context
}

func NewSDKParams(pmk params.ParamsKeeper, ctx sdk.Context) *SDKParams {
	return &SDKParams{
		pmk: pmk,
		ctx: ctx,
	}
}

// The key has the format <module>:(<realm>:)?<paramname>.
func (prm *SDKParams) SetString(key string, value string) {
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetString(prm.ctx, key, value)
}

func (prm *SDKParams) SetBool(key string, value bool) {
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetBool(prm.ctx, key, value)
}

func (prm *SDKParams) SetInt64(key string, value int64) {
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetInt64(prm.ctx, key, value)
}

func (prm *SDKParams) SetUint64(key string, value uint64) {
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetUint64(prm.ctx, key, value)
}

func (prm *SDKParams) SetBytes(key string, value []byte) {
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetBytes(prm.ctx, key, value)
}

func (prm *SDKParams) SetStrings(key string, value []string) {
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetStrings(prm.ctx, key, value)
}

func (prm *SDKParams) willSetKeeperParams(ctx sdk.Context, key string, value interface{}) {
	parts := strings.Split(key, ":")
	if len(parts) == 0 {
		panic(fmt.Sprintf("SDKParams encountered invalid param key format: %s", key))
	}
	mname := parts[0]
	if !prm.pmk.IsRegistered(mname) {
		panic(fmt.Sprintf("module name <%s> not registered", mname))
	}
	kpr := prm.pmk.GetRegisteredKeeper(mname)
	if kpr != nil {
		subkey := key[len(mname)+1:]
		kpr.WillSetParam(prm.ctx, subkey, value)
	}
}
