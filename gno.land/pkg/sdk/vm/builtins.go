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

// GetLockedCoins returns the coins addr's vesting schedule keeps locked at
// the current block time. The total balance, locked coins included, stays
// on GetCoins. No account and no schedule both answer with nothing.
func (bnk *SDKBanker) GetLockedCoins(b32addr crypto.Bech32Address) (dst std.Coins) {
	addr := crypto.MustAddressFromString(string(b32addr))
	acc := bnk.vmk.acck.GetAccount(bnk.ctx, addr)
	if acc == nil {
		return nil
	}
	return acc.LockedCoins(bnk.ctx.BlockTime())
}

// GetCoin reads one denom without touching any other — the O(1) accessor the
// balance split exists to provide. GetCoins costs O(denoms held) and an address
// does not control how many denoms it has been sent.
func (bnk *SDKBanker) GetCoin(b32addr crypto.Bech32Address, denom string) int64 {
	// Every other denom entry point validates — IssueCoin and RemoveCoin via
	// assertCoinDenom, transfers and fees via Coins.IsValid. This one takes a
	// realm-supplied string straight from Gno, so without this it is the only
	// unbounded, unvalidated input reaching a store key. Panic rather than return
	// zero, to match how a malformed denom is treated everywhere else.
	if err := std.ValidateDenom(denom); err != nil {
		panic(err)
	}
	addr := crypto.MustAddressFromString(string(b32addr))
	return bnk.vmk.bank.GetCoin(bnk.ctx, addr, denom)
}

func (bnk *SDKBanker) SendCoins(b32from, b32to crypto.Bech32Address, amt std.Coins) {
	from := crypto.MustAddressFromString(string(b32from))
	// No package account holds a public key, so a send out of an account with
	// one comes from a run's banker kept past its message.
	if acc := bnk.vmk.acck.GetAccount(bnk.ctx, from); acc != nil && acc.GetPubKey() != nil {
		panic("a banker can spend a user's account only inside that user's run")
	}
	bnk.send(from, crypto.MustAddressFromString(string(b32to)), amt)
}

func (bnk *SDKBanker) send(from, to crypto.Address, amt std.Coins) {
	err := bnk.vmk.bank.SendCoins(bnk.ctx, from, to, amt)
	if err != nil {
		panic(err)
	}
}

// RunSDKBanker is the banker of a MsgRun: it alone may spend the run's
// caller, and it refuses any send that takes the coins moved out of from,
// over the whole run, past maxSend, unless the run is uncapped.
type RunSDKBanker struct {
	*SDKBanker
	from crypto.Address
	left map[string]int64 // per denom, what maxSend still allows; nil if uncapped
}

func NewRunSDKBanker(vmk *VMKeeper, ctx sdk.Context, from crypto.Address, maxSend std.Coins, uncapped bool) *RunSDKBanker {
	bnk := &RunSDKBanker{SDKBanker: NewSDKBanker(vmk, ctx), from: from}
	if !uncapped {
		bnk.left = make(map[string]int64, len(maxSend))
		for _, c := range maxSend {
			bnk.left[c.Denom] = c.Amount
		}
	}
	return bnk
}

func (bnk *RunSDKBanker) SendCoins(b32from, b32to crypto.Bech32Address, amt std.Coins) {
	from := crypto.MustAddressFromString(string(b32from))
	if from != bnk.from {
		bnk.SDKBanker.SendCoins(b32from, b32to, amt)
		return
	}
	to := crypto.MustAddressFromString(string(b32to))
	// A zero send moves nothing, and the bank does not validate its denom.
	counted := bnk.left != nil && to != from && !amt.IsZero()
	if counted {
		// Refuse what the bank's SubtractCoins refuses before counting: once amt
		// is sorted, unique and positive, each denom meets its budget once and
		// no coin can raise it. Validate's error names at most two denoms, each
		// within MaxDenomLength.
		if err := amt.Validate(); err != nil {
			panic(std.ErrInvalidCoins(err.Error()))
		}
		for _, c := range amt {
			if c.Amount > bnk.left[c.Denom] {
				panic("run sends more coins than its max send allows")
			}
			bnk.left[c.Denom] -= c.Amount
		}
	}
	bnk.send(from, to, amt)
}

func (bnk *SDKBanker) TotalCoin(denom string) int64 {
	// Validated for the reason GetCoin is: this is a realm-supplied string that
	// reaches a store key, and nothing on the .gno side bounds it.
	if err := std.ValidateDenom(denom); err != nil {
		panic(err)
	}
	return bnk.vmk.bank.TotalSupply(bnk.ctx, denom)
}

// assertIssuable re-checks in Go that denom is realm-qualified.
//
// chain/banker's assertCoinDenom already enforces the full
// "/" + pkgPath + ":" + base shape, but it lives in interpreted stdlib source
// and neither X_bankerIssueCoin nor this type validated anything. That was
// tolerable while the prefix was only a naming convention. It is not now: a bare
// denom reaching here would let a realm mint the chain's own gas denom.
//
// This deliberately checks less than assertCoinDenom does. It answers "is this
// denom realm-qualified at all", not "is it *this* realm's" — the latter stays
// enforced only in .gno, because SDKBanker holds a keeper and a context and not
// the calling realm's package path, which is a machine-frame property. So one
// realm minting another's denom is still blocked one layer up, and only there.
// What is backstopped in Go is the case with a privilege boundary behind it:
// minting outside the realm-denom namespace entirely.
//
// Note this is not the storage-tier question — that is an allowlist in the bank
// (ViewKeeper.inAccountTier), and the two must not be conflated.
func assertIssuable(denom string) {
	if !std.IsRealmDenom(denom) {
		panic("cannot issue or remove a non realm-qualified denom: " + denom)
	}
}

func (bnk *SDKBanker) IssueCoin(b32addr crypto.Bech32Address, denom string, amount int64) {
	assertIssuable(denom)
	addr := crypto.MustAddressFromString(string(b32addr))
	err := bnk.vmk.bank.MintCoins(bnk.ctx, addr, std.Coins{std.Coin{Denom: denom, Amount: amount}})
	if err != nil {
		panic(err)
	}
}

func (bnk *SDKBanker) RemoveCoin(b32addr crypto.Bech32Address, denom string, amount int64) {
	assertIssuable(denom)
	addr := crypto.MustAddressFromString(string(b32addr))
	err := bnk.vmk.bank.BurnCoins(bnk.ctx, addr, std.Coins{std.Coin{Denom: denom, Amount: amount}})
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

// The key has the format <module>:(<realm>:)?<paramname>. Each Set
// wraps in setWithCheck for module-prefix validation, then captures
// the byte delta returned by the params keeper and feeds it to
// recordParamsDelta for per-realm storage-deposit accounting. The
// keeper's Set methods do their own internal Get-with-nil-gctx to
// compute the delta — no separate metered read is needed here.
// See gno.land/pkg/sdk/vm/params_deposit.go.
func (prm *SDKParams) SetString(key string, value string) {
	prm.setWithCheck(key, func() {
		diff := prm.pmk.SetString(prm.ctx, key, value)
		recordParamsDelta(prm.ctx, prm.pmk, key, diff)
	})
}

func (prm *SDKParams) SetBool(key string, value bool) {
	prm.setWithCheck(key, func() {
		diff := prm.pmk.SetBool(prm.ctx, key, value)
		recordParamsDelta(prm.ctx, prm.pmk, key, diff)
	})
}

func (prm *SDKParams) SetInt64(key string, value int64) {
	prm.setWithCheck(key, func() {
		diff := prm.pmk.SetInt64(prm.ctx, key, value)
		recordParamsDelta(prm.ctx, prm.pmk, key, diff)
	})
}

func (prm *SDKParams) SetUint64(key string, value uint64) {
	prm.setWithCheck(key, func() {
		diff := prm.pmk.SetUint64(prm.ctx, key, value)
		recordParamsDelta(prm.ctx, prm.pmk, key, diff)
	})
}

func (prm *SDKParams) SetBytes(key string, value []byte) {
	prm.setWithCheck(key, func() {
		diff := prm.pmk.SetBytes(prm.ctx, key, value)
		recordParamsDelta(prm.ctx, prm.pmk, key, diff)
	})
}

func (prm *SDKParams) SetStrings(key string, value []string) {
	prm.setWithCheck(key, func() {
		diff := prm.pmk.SetStrings(prm.ctx, key, value)
		recordParamsDelta(prm.ctx, prm.pmk, key, diff)
	})
}

// ParamsInterface read methods (G1). Reads do not require module
// registration since they're idempotent and read-only — but they do
// require a colon-prefix to keep the API symmetric with writes.

func (prm *SDKParams) GetString(key string, ptr *string) bool {
	return prm.pmk.GetString(prm.ctx, key, ptr)
}

func (prm *SDKParams) GetBool(key string, ptr *bool) bool {
	return prm.pmk.GetBool(prm.ctx, key, ptr)
}

func (prm *SDKParams) GetInt64(key string, ptr *int64) bool {
	return prm.pmk.GetInt64(prm.ctx, key, ptr)
}

func (prm *SDKParams) GetUint64(key string, ptr *uint64) bool {
	return prm.pmk.GetUint64(prm.ctx, key, ptr)
}

func (prm *SDKParams) GetBytes(key string, ptr *[]byte) bool {
	return prm.pmk.GetBytes(prm.ctx, key, ptr)
}

func (prm *SDKParams) GetStrings(key string, ptr *[]string) bool {
	return prm.pmk.GetStrings(prm.ctx, key, ptr)
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
