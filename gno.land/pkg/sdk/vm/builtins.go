package vm

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
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
	pmk ParamsKeeperI
	ctx sdk.Context
}

func NewSDKParams(pmk ParamsKeeperI, ctx sdk.Context) *SDKParams {
	return &SDKParams{
		pmk: pmk,
		ctx: ctx,
	}
}

// The key has the format <module>:(<realm>:)?<paramname>.
func (prm *SDKParams) SetString(key string, value string) {
	prm.setWithCheck(key, func() { prm.pmk.SetString(prm.ctx, key, value) })
}

func (prm *SDKParams) SetBool(key string, value bool) {
	prm.setWithCheck(key, func() { prm.pmk.SetBool(prm.ctx, key, value) })
}

func (prm *SDKParams) SetInt64(key string, value int64) {
	prm.setWithCheck(key, func() { prm.pmk.SetInt64(prm.ctx, key, value) })
}

func (prm *SDKParams) SetUint64(key string, value uint64) {
	prm.setWithCheck(key, func() { prm.pmk.SetUint64(prm.ctx, key, value) })
}

func (prm *SDKParams) SetBytes(key string, value []byte) {
	prm.setWithCheck(key, func() { prm.pmk.SetBytes(prm.ctx, key, value) })
}

func (prm *SDKParams) SetStrings(key string, value []string) {
	prm.setWithCheck(key, func() { prm.pmk.SetStrings(prm.ctx, key, value) })
}

func (prm *SDKParams) UpdateStrings(key string, vals []string, add bool) {
	prm.mustHaveModuleKeeper(key)
	ss := &[]string{}
	prm.pmk.GetStrings(prm.ctx, key, ss)

	oldList := *ss
	existing := make(map[string]struct{}, len(oldList))
	// Temporary map for duplicate detection
	for _, s := range oldList {
		existing[s] = struct{}{}
	}

	if add {
		// Append only non-duplicate values
		for _, v := range vals {
			if _, found := existing[v]; !found {
				oldList = append(oldList, v)
				existing[v] = struct{}{}
			}
		}
		prm.SetStrings(key, oldList)
		return
	}
	// Remove case
	updatedList := oldList[:0] // reuse original memory
	removeSet := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		removeSet[v] = struct{}{}
	}

	for _, s := range oldList {
		if _, found := removeSet[s]; !found {
			updatedList = append(updatedList, s)
		}
	}
	prm.SetStrings(key, updatedList)
}

func (prm *SDKParams) setWithCheck(key string, set func()) {
	prm.mustHaveModuleKeeper(key)
	set()
}

func (prm *SDKParams) mustHaveModuleKeeper(key string) {
	idx := strings.Index(key, ":")
	if idx <= 0 {
		panic(fmt.Sprintf("SDKParams encountered invalid param key format: %s", key))
	}
	mname := key[:idx]
	if !prm.pmk.IsRegistered(mname) {
		panic(fmt.Sprintf("module name <%s> not registered", mname))
	}
}
