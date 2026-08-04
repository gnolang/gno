package bank

import (
	"bytes"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// SupplyPrefix is where the per-denom supply counter lives:
//
//	/supply/ || denom   ->   amount, 8-byte big-endian
//
// Spelled out rather than abbreviated. Every module shares one store here, so the
// prefix is what namespaces the keyspace — cosmos can afford a single opaque byte
// because each of its modules gets its own mounted store, and we cannot. Store keys
// are not gas-metered, so the extra bytes cost nothing but state.
//
// Deliberately not "/s/": that is auth.SessionStoreKeyInfix verbatim, and while a
// session key begins "/a/" so the two ranges cannot overlap, sharing the spelling
// would make a search for one return the other.
const SupplyPrefix = "/supply/"

// maxSupplyDenoms bounds the accumulator the *invariant* uses to total balances.
// Genesis seeding is deliberately unbounded; see computeSupply.
//
// At a maximal 274-byte denom this is roughly 80 MB, not the 8 MB the account-number
// bitset costs at its own bound — the shapes differ, only the reasoning is shared.
//
// Distinct denoms are attacker-driven and unbounded — every IssueCoin to a fresh
// realm denom creates one — and an out-of-memory abort is fatal, so no guard could
// recover it. Above this bound the caller reports what it could not verify rather
// than allocating. Same reasoning as auth.maxUniquenessBits.
const maxSupplyDenoms = 1 << 18

// SupplyKey returns the store key holding denom's total supply.
func SupplyKey(denom string) []byte {
	return append([]byte(SupplyPrefix), denom...)
}

// denomFromSupplyKey recovers the denom from a supply key. Non-panicking: the
// invariants sweep the whole keyspace and must report a malformed key, not die.
func denomFromSupplyKey(key []byte) (string, error) {
	if !bytes.HasPrefix(key, []byte(SupplyPrefix)) {
		return "", fmt.Errorf("key %X is not under %q", key, SupplyPrefix)
	}
	denom := string(key[len(SupplyPrefix):])
	if denom == "" {
		return "", fmt.Errorf("malformed supply key: %X has no denom", key)
	}
	return denom, nil
}

// TotalSupply returns how much of denom exists, or zero for a denom nobody holds.
func (view ViewKeeper) TotalSupply(ctx sdk.Context, denom string) int64 {
	// Bounded before anything hashes or copies it, for the reason GetCoin is; a
	// denom this long cannot have a supply record, so zero is exact.
	if len(denom) > std.MaxDenomLength {
		return 0
	}
	bz := ctx.Store(view.key).Get(ctx.GasContext(), SupplyKey(denom))
	if bz == nil {
		return 0
	}
	return decodeBalance(bz)
}

// setSupply writes denom's supply, deleting the key when it reaches zero.
//
// Deleting rather than storing a zero matches balances, and for the same reason:
// encodeBalance refuses a non-positive amount, so a stored zero would be reported
// as corruption by machinery that already exists. It also means a fully burned
// denom leaves no trace, and TotalSupply cannot distinguish "never existed" from
// "all burned" — which is the same answer either way.
func (bank BankKeeper) setSupply(ctx sdk.Context, denom string, amount int64) {
	stor := ctx.Store(bank.key)
	key := SupplyKey(denom)
	if amount == 0 {
		stor.Delete(ctx.GasContext(), key)
		return
	}
	stor.Set(ctx.GasContext(), key, encodeBalance(amount))
}

// nextSupply computes each denom's supply after applying amt with the given sign,
// without writing anything. Reports an error rather than panicking: overflow is
// reachable by an ordinary realm minting too much, which is a caller's mistake
// rather than a broken node.
func (bank BankKeeper) nextSupply(ctx sdk.Context, amt std.Coins, sign int64) (std.Coins, error) {
	next := make(std.Coins, len(amt))
	for i, coin := range amt {
		if i > 0 && amt[i-1].Denom >= coin.Denom {
			// Two entries for one denom would each compute from the same old value
			// and the last write would lose an increment. Callers validate, so this
			// cannot fire today; the arithmetic below depends on it, so it is checked
			// here rather than inherited.
			return nil, fmt.Errorf("denoms out of order or duplicated: %q then %q",
				amt[i-1].Denom, coin.Denom)
		}
		old := bank.TotalSupply(ctx, coin.Denom)
		sum, ok := overflow.Add(old, sign*coin.Amount)
		if !ok || sum < 0 {
			return nil, std.ErrInvalidCoins(fmt.Sprintf(
				"supply of %s would go from %d to %d%+d, which is out of range",
				coin.Denom, old, old, sign*coin.Amount))
		}
		next[i] = std.Coin{Denom: coin.Denom, Amount: sum}
	}
	return next, nil
}

// MintCoins creates amt and credits it to addr, raising the supply counter.
//
// This and BurnCoins are the only paths that change supply. Transfers are
// supply-neutral, so AddCoins and subtract deliberately do not touch the counter:
// they also carry fees and storage-deposit refunds, and a counter that followed
// them would make an unpaired credit look legitimate instead of breaking the
// supply invariant — which is the whole point of keeping a second number.
//
// Supply is capped at MaxInt64 per denom. That is a real constraint, not a
// formality: before this counter existed two addresses could each hold MaxInt64 of
// one denom, since AddCoins bounds each balance and nothing bounded the total.
func (bank BankKeeper) MintCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}
	// Everything that can fail happens before anything is written, so a rejected
	// mint leaves neither the balance nor the counter changed.
	next, err := bank.nextSupply(ctx, amt, 1)
	if err != nil {
		return err
	}
	if err := bank.AddCoins(ctx, addr, amt); err != nil {
		return err
	}
	for _, coin := range next {
		bank.setSupply(ctx, coin.Denom, coin.Amount)
	}
	return nil
}

// BurnCoins destroys amt held by addr, lowering the supply counter.
func (bank BankKeeper) BurnCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}
	// Checked before the debit, so a burn of more than was ever minted — which
	// means the counter already disagreed with the balances — cannot make it worse.
	next, err := bank.nextSupply(ctx, amt, -1)
	if err != nil {
		return err
	}
	if err := bank.SubtractCoins(ctx, addr, amt); err != nil {
		return err
	}
	for _, coin := range next {
		bank.setSupply(ctx, coin.Denom, coin.Amount)
	}
	return nil
}

// computeSupply totals every balance, across both tiers, per denom.
//
// Shared by RecomputeSupply and the supply invariant so that the recorded number
// and the checked number are produced by one piece of code. Reads raw: the typed
// accessors panic on states this must report or survive.
// maxDenoms of 0 means unlimited. The invariant passes a bound because it must not
// risk an unrecoverable allocation on a node; genesis passes none, because a fork of a
// chain that legitimately holds many denoms has to be able to re-seed, and refusing to
// start would be worse than the memory.
func computeSupply(ctx sdk.Context, view ViewKeeper, maxDenoms int) (map[string]int64, error) {
	totals := make(map[string]int64)
	capped := false
	add := func(denom string, amount int64) error {
		if _, ok := totals[denom]; !ok {
			if maxDenoms > 0 && len(totals) >= maxDenoms {
				capped = true
				return nil
			}
		}
		sum, ok := overflow.Add(totals[denom], amount)
		if !ok {
			return fmt.Errorf("balances of %q sum past int64", denom)
		}
		totals[denom] = sum
		return nil
	}

	iter := store.PrefixIterator(nil, ctx.Store(view.key), []byte(BalancePrefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, denom, err := parseBalanceKey(iter.Key())
		if err != nil {
			return nil, err
		}
		amount, err := tryDecodeBalance(iter.Value())
		if err != nil {
			return nil, err
		}
		if err := add(denom, amount); err != nil {
			return nil, err
		}
	}
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iteration over %q failed: %w", BalancePrefix, err)
	}

	// The account tier is summed raw and unfiltered. A denom stranded there by an
	// allowlist change is still stored value that the numbers must account for, and
	// accountTierCoins would panic on exactly that state.
	// addErr is carried out of the callback deliberately. IterateAccountEntries
	// returns the *iterator's* error, so returning true from the callback only stops
	// the walk — the error would be lost and computeSupply would return truncated
	// totals with err == nil, which RecomputeSupply would seed and SupplyInvariant
	// would then agree with.
	var addErr error
	err := view.acck.IterateAccountEntries(ctx, func(e auth.AccountEntry) bool {
		if e.Kind != auth.AccountKeyRegular || e.DecodeErr != nil || e.Account == nil {
			return false
		}
		for _, coin := range e.Account.GetCoins() {
			if addErr = add(coin.Denom, coin.Amount); addErr != nil {
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, fmt.Errorf("iteration over accounts failed: %w", err)
	}
	if addErr != nil {
		return nil, addErr
	}
	if capped {
		return nil, fmt.Errorf("more than %d distinct denoms", maxDenoms)
	}
	return totals, nil
}

// RecomputeSupply rewrites every supply record from the balances actually held.
//
// For genesis and offline tooling only, never a metered path and never a live
// chain: it writes state outside any block, so a node that ran it and one that did
// not would diverge.
//
// Genesis needs this rather than an incremental hook because SetCoins cannot
// compute its own delta. InitChainer writes the account object with the full
// pre-split balance *before* calling SetCoins, so SetCoins reads old == new and a
// delta would be zero for every vesting account — and that pre-write cannot be
// removed, since the vesting constructors validate OriginalVesting against it.
func (bank BankKeeper) RecomputeSupply(ctx sdk.Context) {
	totals, err := computeSupply(ctx, bank.ViewKeeper, 0)
	if err != nil {
		panic(fmt.Sprintf("cannot compute supply: %v", err))
	}
	// Clear first: a denom that no longer exists must not keep a stale record.
	stor := ctx.Store(bank.key)
	iter := store.PrefixIterator(nil, stor, []byte(SupplyPrefix))
	var stale [][]byte
	for ; iter.Valid(); iter.Next() {
		stale = append(stale, append([]byte(nil), iter.Key()...))
	}
	iterErr := iter.Error()
	iter.Close()
	if iterErr != nil {
		panic(fmt.Sprintf("cannot read existing supply records: %v", iterErr))
	}
	for _, key := range stale {
		stor.Delete(nil, key)
	}
	// Ranging a map to write is safe here only because the cache store sorts its keys
	// before flushing to the committed tree, so the tree's shape does not depend on
	// Go's map order. Verified: 25 identical genesis runs produce one app hash.
	for denom, amount := range totals {
		if amount != 0 {
			stor.Set(nil, SupplyKey(denom), encodeBalance(amount))
		}
	}
}
