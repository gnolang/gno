package vm

import (
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
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

func (prm *SDKParams) SetString(key, value string) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetString(prm.ctx, key, value)
}

// Set a boolean parameter in the format of realmPath.parameter.bool
func (prm *SDKParams) SetBool(key string, value bool) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetBool(prm.ctx, key, value)
}

func (prm *SDKParams) SetInt64(key string, value int64) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetInt64(prm.ctx, key, value)
}

func (prm *SDKParams) SetUint64(key string, value uint64) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetUint64(prm.ctx, key, value)
}

func (prm *SDKParams) SetBytes(key string, value []byte) {
	prm.assertRealmAccess(key)
	prm.willSetKeeperParams(prm.ctx, key, value)
	prm.pmk.SetBytes(prm.ctx, key, value)
}

// willSetKeeperParams parses the parameter key and sets the keeper it matches the keeper key
// For the system params, the internal key format is sysParamsRealm.[keeperKeyPrefix:]keyName.keyType
// Ex. gno.lang/r/sys/params.bank:lockStrn.string
// The "keeperKeyPrefix:" is optional.
// If "keeperKeyPrefix:" is presented in the key,
// it must match the keeper's paramKeyPrefix; otherwise it will panic and revert the transaction.
func (prm *SDKParams) willSetKeeperParams(ctx sdk.Context, key string, value interface{}) {
	// key is in the format of <realm>:<keyname>.<type>
	realmPrefix := gno.ReRealmPath.FindString(key)
	if realmPrefix == "" {
		panic(fmt.Sprintf("set parameter %s must be accessed from a realm.", key))
	}
	// ParamfulKeeper can be accessed from SysParamsRealmPath only
	if realmPrefix != SysParamsRealmPath {
		return
	}

	k, ok := strings.CutPrefix(key, realmPrefix)
	if !ok {
		return
	}

	parts := strings.SplitN(k, ".", 2)
	paramKey := parts[1]
	parts = strings.SplitN(paramKey, ":", 2)
	keeperKeyPrefix := ""

	if len(parts) == 2 {
		keeperKeyPrefix = parts[0]
		paramKey = parts[1]
	} else {
		// no keeperKeyPrefix
		return
	}

	if !prm.pmk.IsRegistered(keeperKeyPrefix) {
		panic(fmt.Sprintf("keeper key <%s> does not exist", keeperKeyPrefix))
	}
	kpr := prm.pmk.GetRegisteredKeeper(keeperKeyPrefix)
	if kpr != nil {
		kpr.WillSetParam(prm.ctx, paramKey, value)
	}
}

func (prm *SDKParams) assertRealmAccess(key string) {
	realm := gno.ReRealmPath.FindString(key)
	if realm == "" {
		panic(fmt.Sprintf("Set parameters must be accessed from a realm"))
	}
}
