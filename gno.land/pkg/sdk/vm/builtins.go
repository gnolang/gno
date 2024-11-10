package vm

import (
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

type SDKParams struct {
	vmk *VMKeeper
	ctx sdk.Context
}

func NewSDKParams(vmk *VMKeeper, ctx sdk.Context) *SDKParams {
	return &SDKParams{
		vmk: vmk,
		ctx: ctx,
	}
}

func (prm *SDKParams) SetString(key, value string) {
	if err := prm.vmk.prmk.SetString(prm.ctx, key, value); err != nil {
		panic(err)
	}
}

func (prm *SDKParams) SetBool(key string, value bool) {
	if err := prm.vmk.prmk.SetBool(prm.ctx, key, value); err != nil {
		panic(err)
	}
}

func (prm *SDKParams) SetInt64(key string, value int64) {
	if err := prm.vmk.prmk.SetInt64(prm.ctx, key, value); err != nil {
		panic(err)
	}
}

func (prm *SDKParams) SetUint64(key string, value uint64) {
	if err := prm.vmk.prmk.SetUint64(prm.ctx, key, value); err != nil {
		panic(err)
	}
}

func (prm *SDKParams) SetBytes(key string, value []byte) {
	if err := prm.vmk.prmk.SetBytes(prm.ctx, key, value); err != nil {
		panic(err)
	}
}
