package bank

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Balance storage.
//
// A balance whose denom is in the account tier lives inside the account object,
// in its Coins field. Every other balance lives on its own at
//
//	/b/ || addr (20 raw bytes) || denom      ->  amount, 8-byte big-endian
//
// The account tier is an allowlist of the chain's gas denoms, injected at
// construction (see NewViewKeeper). It is tiny by design: a balance there is free
// only because the account object is rewritten on every transaction its owner
// sends anyway, to bump the sequence, and that argument holds for gas denoms and
// nothing else. Anything else kept there would be unbounded
// third-party-writable data on the holder's own metered critical path — see the
// ADR for the attack that enables, and for why the allowlist is closed and
// injected rather than a governance param.
//
// Not to be confused with who may *issue* a denom, which is std.IsRealmDenom and
// is enforced at the banker.
//
// Layout constraints, all load-bearing:
//   - The address is not length-prefixed, which is safe only while
//     crypto.Address is a fixed [20]byte: the denom is the remainder at a
//     constant offset, so it may itself contain "/" and ":" freely.
//   - Store keys are not gas-metered, so key size is bounded only by
//     MaxDenomLength (274) in tm2/pkg/std. Value size is metered, which is why
//     the amount is fixed-width rather than amino-encoded.
//   - "/b/" must not prefix, or be prefixed by, any other key in the shared main
//     store. Check bare names too — "gasPrice", "globalAccountNumber", "pkg:",
//     "consensus_params", GnoVM object ids — not only "/..." prefixes. Pinned by
//     TestBalancePrefixDoesNotCollide. Deliberately not nested under "/a/",
//     where the session-account precedent shows nesting creates an
//     iteration-filter hazard.
const (
	// BalancePrefix namespaces split-tier balances — every denom outside the
	// account-tier allowlist, which is a strict superset of the realm-issued ones.
	BalancePrefix = "/b/"

	// balanceAmountLen is the encoded width of an amount.
	balanceAmountLen = 8
)

// BalanceKey returns the store key holding addr's balance of denom.
//
// Only valid for split-tier denoms; account-tier denoms live in the account
// object.
func BalanceKey(addr crypto.Address, denom string) []byte {
	key := make([]byte, 0, len(BalancePrefix)+crypto.AddressSize+len(denom))
	key = append(key, BalancePrefix...)
	key = append(key, addr[:]...)
	key = append(key, denom...)
	return key
}

// BalancePrefixKey returns the prefix covering every balance held by addr, the
// analogue of auth.SessionPrefixKey. Iterating it yields denoms in ascending order.
func BalancePrefixKey(addr crypto.Address) []byte {
	key := make([]byte, 0, len(BalancePrefix)+crypto.AddressSize)
	key = append(key, BalancePrefix...)
	key = append(key, addr[:]...)
	return key
}

// denomFromBalanceKey recovers the denom from a full balance key. The address is
// fixed-width, so this is a constant offset.
func denomFromBalanceKey(key []byte) (string, error) {
	_, denom, err := parseBalanceKey(key)
	return denom, err
}

// parseBalanceKey splits a full balance key into its address and denom.
//
// Unlike denomFromBalanceKey this also checks the prefix, because a caller that
// did not get the key from an iterator bounded by BalancePrefix has no other
// guarantee it is a balance key at all. That is the case for the invariants,
// which sweep the whole keyspace.
func parseBalanceKey(key []byte) (crypto.Address, string, error) {
	var addr crypto.Address
	if !bytes.HasPrefix(key, []byte(BalancePrefix)) {
		return addr, "", fmt.Errorf("key %X is not under %q", key, BalancePrefix)
	}
	prefixLen := len(BalancePrefix) + crypto.AddressSize
	if len(key) <= prefixLen {
		return addr, "", fmt.Errorf("malformed balance key: %X", key)
	}
	copy(addr[:], key[len(BalancePrefix):prefixLen])
	return addr, string(key[prefixLen:]), nil
}

// encodeBalance encodes a positive amount as fixed-width big-endian.
//
// Fixed width rather than amino: amino encodes int64(0) as zero bytes, and the
// store panics on a nil value, so an accidentally-persisted zero would be a
// panic rather than a caught bug. Fixed width also makes the write cost
// independent of the balance's magnitude.
func encodeBalance(amount int64) []byte {
	if amount <= 0 {
		// Callers must delete instead; a stored balance is always positive.
		panic(fmt.Sprintf("refusing to encode non-positive balance %d", amount))
	}
	bz := make([]byte, balanceAmountLen)
	binary.BigEndian.PutUint64(bz, uint64(amount))
	return bz
}

// decodeBalance decodes a stored amount. A stored value is always a positive
// int64, so anything else means state corruption.
func decodeBalance(bz []byte) int64 {
	amount, err := tryDecodeBalance(bz)
	if err != nil {
		// panic with the string, not the error: the recovered value's type is
		// observable, and tests assert on it.
		panic(err.Error())
	}
	return amount
}

// tryDecodeBalance is decodeBalance without the panic.
//
// The money path must panic — a corrupt balance must never be silently treated as
// a number. The invariants are the one caller that has to survive the corruption
// they are looking for, and must report every instance rather than stop at the
// first. Keeping both on one implementation means the wire format has a single
// source of truth.
func tryDecodeBalance(bz []byte) (int64, error) {
	if len(bz) != balanceAmountLen {
		return 0, fmt.Errorf("corrupt balance value: expected %d bytes, got %d",
			balanceAmountLen, len(bz))
	}
	u := binary.BigEndian.Uint64(bz)
	if u == 0 || u > uint64(math.MaxInt64) {
		return 0, fmt.Errorf("corrupt balance value: %d out of range", u)
	}
	return int64(u), nil
}

// getSplitBalance returns addr's balance of a split-tier denom, or zero.
func (view ViewKeeper) getSplitBalance(ctx sdk.Context, addr crypto.Address, denom string) int64 {
	bz := ctx.Store(view.key).Get(ctx.GasContext(), BalanceKey(addr, denom))
	if bz == nil {
		return 0
	}
	return decodeBalance(bz)
}

// setSplitBalance writes addr's balance of a split-tier denom, deleting the key
// when the balance reaches zero.
//
// Zero balances are never persisted. That keeps the reconstructed Coins valid
// (std.Coins rejects non-positive amounts) and stops drained denoms
// accumulating as dead keys forever.
func (bank BankKeeper) setSplitBalance(ctx sdk.Context, addr crypto.Address, denom string, amount int64) {
	stor := ctx.Store(bank.key)
	key := BalanceKey(addr, denom)
	if amount == 0 {
		stor.Delete(ctx.GasContext(), key)
		return
	}
	stor.Set(ctx.GasContext(), key, encodeBalance(amount))
}

// splitCoins returns every split-tier balance held by addr, in ascending denom
// order.
//
// O(number of split-tier denoms held), and the one operation the split does not
// make cheap. No transfer reaches it: the callers are the ones that genuinely
// want the whole set — the balances query, Gno's banker.GetCoins — plus SetCoins,
// which enumerates to find the stale keys it must delete.
func (view ViewKeeper) splitCoins(ctx sdk.Context, addr crypto.Address) std.Coins {
	var coins std.Coins
	iter := store.PrefixIterator(ctx.GasContext(), ctx.Store(view.key), BalancePrefixKey(addr))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom, err := denomFromBalanceKey(iter.Key())
		if err != nil {
			panic(err)
		}
		// Ascending order is what lets GetCoins merge without revalidating. The
		// store guarantees it; assert it rather than trust it, since a violation
		// would silently produce invalid std.Coins.
		if n := len(coins); n > 0 && coins[n-1].Denom >= denom {
			panic(fmt.Sprintf("balance keys out of order: %q then %q", coins[n-1].Denom, denom))
		}
		// A denom must live in exactly one tier. If one is ever moved into the
		// account tier without migrating its existing key, both would hold it and
		// GetCoins would silently sum them — reporting a balance GetCoin
		// disagrees with. Fail loudly at the first read instead. This is the
		// invariant that makes the allowlist a construction argument rather than
		// a governance param; see the ADR.
		if view.inAccountTier(denom) {
			panic(fmt.Sprintf(
				"denom %q has a split-tier key but is in the account tier: "+
					"the allowlist changed without migrating existing balances", denom))
		}
		coins = append(coins, std.Coin{Denom: denom, Amount: decodeBalance(iter.Value())})
	}
	return coins
}
