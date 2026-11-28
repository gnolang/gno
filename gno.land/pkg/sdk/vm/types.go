package vm

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// AccountKeeperI is the limited interface only needed for VM.
type AccountKeeperI interface {
	GetAccount(ctx sdk.Context, addr crypto.Address) std.Account
	SetSessionAccount(ctx sdk.Context, master crypto.Address, acc std.Account)
}

// BankKeeperI is the limited interface only needed for VM.
type BankKeeperI interface {
	GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins
	GetCoin(ctx sdk.Context, addr crypto.Address, denom string) int64
	SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	// Here for the keeper's own storage-deposit refund, which pays out of the
	// realm's escrow and must bypass the transfer restriction, and for the
	// charge while genesis replays history. No native reaches
	// it — SDKBanker.SendCoins uses the restricted SendCoins — so a realm cannot
	// spend through it. Keep it that way: it is the one method here that skips
	// every check.
	SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	// Here for the keeper's own storage-deposit charge. It enforces the caller's
	// vesting schedule because the refund goes to whoever frees the storage. No
	// native reaches it either.
	SendCoinsVestingChecked(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	// Issuance goes through Mint/Burn, which maintain the supply counter. The raw
	// AddCoins/SubtractCoins are deliberately absent: they are supply-blind, so
	// exposing them here would make an unaccounted mint expressible from a realm.
	MintCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	BurnCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	TotalSupply(ctx sdk.Context, denom string) int64
	RestrictedDenoms(ctx sdk.Context) []string
}

// ParamsKeeperI is the limited interface only needed for VM.
type ParamsKeeperI interface {
	params.ParamsKeeperI

	IsRegistered(moduleName string) bool
}

// Public facing function signatures.
// See convertArgToGno() for supported types.
type FunctionSignature struct {
	FuncName string
	Params   []NamedType
	Results  []NamedType
}

type NamedType struct {
	Name  string
	Type  string
	Value string
}

type FunctionSignatures []FunctionSignature

func (fsigs FunctionSignatures) JSON() string {
	bz := amino.MustMarshalJSON(fsigs)
	return string(bz)
}
