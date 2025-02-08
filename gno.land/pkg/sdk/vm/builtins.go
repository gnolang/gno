package vm

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gstd "github.com/gnolang/gno/gnovm/stdlibs/std"
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

type SDKParams struct {
	pmk *params.ParamsKeeper
	ctx sdk.Context
}

// These are the native function implementations bound to standard libraries in Gno.
// All methods of this struct are not intended to be called from outside vm/stdlibs/std.
//
// The key has the format <realm>.<paramname>.<type>:
// realm: A realm path indicating where Set methods are called from.
// paramname: The parameter key. If it contains a prefix that matches the module's paramkey
// prefix (which by default is the module name), it indicates an attempt to set the module's
// parameters for the chain. Otherwise, it is treated as an arbitrary parameter.
// type: The primitive type of the parameter value.

func NewSDKParams(pmk *params.ParamsKeeper, ctx sdk.Context) *SDKParams {
	return &SDKParams{
		pmk: pmk,
		ctx: ctx,
	}
}

func (prm *SDKParams) SetString(key gstd.ParamKey, value string) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetString(prm.ctx, key.String(), value)
}

// Set a boolean parameter in the format of realmPath.parameter.bool
func (prm *SDKParams) SetBool(key gstd.ParamKey, value bool) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetBool(prm.ctx, key.String(), value)
}

func (prm *SDKParams) SetInt64(key gstd.ParamKey, value int64) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetInt64(prm.ctx, key.String(), value)
}

func (prm *SDKParams) SetUint64(key gstd.ParamKey, value uint64) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetUint64(prm.ctx, key.String(), value)
}

func (prm *SDKParams) SetBytes(key gstd.ParamKey, value []byte) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetBytes(prm.ctx, key.String(), value)
}

// ParamKey's prefix must match the keeper's paramKeyPrefix; otherwise it will panic and revert the transaction.
func (prm *SDKParams) willSetKeeperParams(ctx sdk.Context, key gstd.ParamKey, value interface{}) {
	kp := key.Prefix
	// ParamfulKeeper can be accessed from SysParamsRealmPath only
	if key.Realm != SysParamsRealmPath || kp == "" {
		return
	}

	if !prm.pmk.IsRegistered(kp) {
		panic(fmt.Sprintf("keeper key <%s> does not exist", kp))
	}
	kpr := prm.pmk.GetRegisteredKeeper(kp)
	if kpr != nil {
		kpr.WillSetParam(prm.ctx, key.Key, value)
	}
}

func (prm *SDKParams) assertRealmAccess(key gstd.ParamKey) {
	realm := gno.ReRealmPath.FindString(key.Realm)
	if realm == "" {
		panic(fmt.Sprintf("parameters must be set in a valid realm"))
	}

	if key.Realm != SysParamsRealmPath && key.Prefix != "" {
		panic(fmt.Sprintf("prefixed parameter %q with keeper prefix %q must be set in %q", key.Key, key.Prefix, SysParamsRealmPath))
	}
}
