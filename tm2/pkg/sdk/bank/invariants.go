package bank

import (
	"bytes"
	"maps"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// ModuleName is used to label findings; see invariant names below.

// BalanceKeysInvariant sweeps the split-tier keyspace.
//
// Checks that every key under BalancePrefix is well formed and names a valid denom;
// that every value decodes to a positive amount; that no denom there belongs to the
// account tier; that a realm-shaped denom has a shape a realm could have issued; and
// that every address holding a balance has an account object, without which the
// funds are permanently unspendable because the address cannot sign.
//
// Reads nothing through the keeper's typed accessors. splitCoins, GetCoins and
// decodeBalance all panic on precisely the states reported here, and an invariant
// must enumerate every violation rather than die at the first.
func BalanceKeysInvariant(view ViewKeeper) sdk.Invariant {
	return sdk.Guard(ModuleName, "balance-keys", func(ctx sdk.Context, rep *sdk.InvariantReport) {
		// nil gas context: a sweep of a whole keyspace under a live meter would
		// run out of gas partway through, and Guard would flatten that into a
		// single "check panicked" line — reporting nothing about the state.
		iter := store.PrefixIterator(nil, ctx.Store(view.key), []byte(BalancePrefix))
		defer iter.Close()

		var prevKey []byte
		var curAddr crypto.Address
		var haveAddr bool
		for ; iter.Valid(); iter.Next() {
			key := iter.Key()

			// Ordering is what lets GetCoins merge the tiers without revalidating;
			// splitCoins panics if it is ever violated. Not producible by stored
			// state — the store guarantees it — so this only catches a store bug.
			if prevKey != nil && bytes.Compare(prevKey, key) >= 0 {
				rep.Addf("keys out of order: %X then %X", prevKey, key)
			}
			prevKey = append(prevKey[:0], key...)

			addr, denom, err := parseBalanceKey(key)
			if err != nil {
				rep.Addf("%v", err)
				continue
			}
			if err := std.ValidateDenom(denom); err != nil {
				rep.Addf("address %s holds an invalid denom: %v", addr, err)
			}
			if _, err := tryDecodeBalance(iter.Value()); err != nil {
				rep.Addf("address %s denom %q: %v", addr, denom, err)
			}
			if view.inAccountTier(denom) {
				rep.Addf("address %s has a split-tier key for account-tier denom %q: "+
					"the allowlist grew without migrating existing balances", addr, denom)
			}
			if std.IsRealmDenom(denom) {
				if _, _, err := std.ParseRealmDenom(denom); err != nil {
					rep.Addf("address %s holds a realm-shaped denom no realm could issue "+
						"(genesis-authored or corrupt): %v", addr, err)
				}
			}
			// Once per address, not per key: the keyspace sorts address-major, so an
			// address's balances are contiguous.
			if !haveAddr || addr != curAddr {
				curAddr, haveAddr = addr, true
				if !view.acck.HasAccount(ctx, addr) {
					rep.Addf("address %s holds a balance but has no account object, "+
						"so it cannot sign and the funds are unspendable", addr)
				}
			}
		}
		// A failed iteration must never read as a healthy keyspace: the bptree store
		// panics from Valid()/Close() when its error goes unread, and the iavl store
		// reports nothing, which would silently truncate the sweep.
		if err := iter.Error(); err != nil {
			rep.Addf("iteration over %q failed, so the sweep is incomplete: %v", BalancePrefix, err)
		}
	})
}

// AccountTierInvariant checks the balances held inside account objects.
//
// Every denom there must be in the allowlist. A denom that is not is stranded: the
// keeper routes it to the split tier, so GetCoin reports zero while GetCoins reports
// it, and it cannot be spent. A realm-issuable denom there is worse than stranded and
// is reported as such — its issuer could mint into a shared metered blob, which is
// the griefing the split exists to close.
func AccountTierInvariant(view ViewKeeper) sdk.Invariant {
	return sdk.Guard(ModuleName, "account-tier", func(ctx sdk.Context, rep *sdk.InvariantReport) {
		stor := ctx.Store(view.key)
		err := view.acck.IterateAccountEntries(ctx, func(e auth.AccountEntry) bool {
			if e.Kind != auth.AccountKeyRegular || e.DecodeErr != nil {
				return false // auth's invariant owns key shape and decodability
			}
			for _, coin := range e.Account.GetCoins() {
				if view.inAccountTier(coin.Denom) {
					continue
				}
				what := "is not in the account-tier allowlist"
				if std.IsRealmDenom(coin.Denom) {
					what = "is realm-issuable and must never be in the account tier"
				}
				extra := ""
				// Only on the failure path, so it costs nothing when healthy.
				if stor.Has(nil, BalanceKey(e.Addr, coin.Denom)) {
					extra = " (it also has a split-tier key: double-homed, so GetCoins would sum both)"
				}
				rep.Addf("address %s account object holds %q, which %s%s", e.Addr, coin.Denom, what, extra)
			}
			// Vesting schedules are read from the concrete type: the interface's
			// GetStartTime returns a hardcoded zero for delayed accounts, so a
			// schedule rebuilt from the getters would be wrong for them.
			if sched, ok := vestingScheduleOf(e.Account); ok {
				if err := sched.Validate(); err != nil {
					rep.Addf("address %s has an invalid vesting schedule: %v", e.Addr, err)
				}
			}
			return false
		})
		if err != nil {
			rep.Addf("iteration over accounts failed, so the sweep is incomplete: %v", err)
		}
	})
}

// vestingScheduleOf returns the stored schedule, reading the concrete type rather
// than the std.VestingAccount interface.
func vestingScheduleOf(acc std.Account) (std.VestingSchedule, bool) {
	switch a := acc.(type) {
	case *std.ContinuousVestingAccount:
		return a.VestingSchedule, true
	case *std.DelayedVestingAccount:
		return a.VestingSchedule, true
	case *std.BaseVestingAccount:
		return a.VestingSchedule, true
	}
	return std.VestingSchedule{}, false
}

// SupplyInvariant checks the recorded supply against the balances actually held.
//
// This is the only redundancy check in the set: two numbers maintained by different
// code that must agree. The others are structural — they check shapes, tiers and
// key/field agreement — and a mint that bypassed the counter looks perfectly well
// formed to all of them. Here it does not.
//
// Reports in both directions. A denom held with no record is what an unaccounted
// credit produces; a record with nothing held is what a lost balance write produces.
func SupplyInvariant(view ViewKeeper) sdk.Invariant {
	return sdk.Guard(ModuleName, "total-supply", func(ctx sdk.Context, rep *sdk.InvariantReport) {
		held, err := computeSupply(ctx, view, maxSupplyDenoms)
		if err != nil {
			rep.Addf("cannot total the balances, so supply was NOT verified: %v", err)
			return
		}
		iter := store.PrefixIterator(nil, ctx.Store(view.key), []byte(SupplyPrefix))
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			denom, err := denomFromSupplyKey(iter.Key())
			if err != nil {
				rep.Addf("%v", err)
				continue
			}
			recorded, err := tryDecodeBalance(iter.Value())
			if err != nil {
				rep.Addf("supply record for %q: %v", denom, err)
				delete(held, denom)
				continue
			}
			if sum, ok := held[denom]; !ok {
				rep.Addf("supply of %q is recorded as %d but no address holds any",
					denom, recorded)
			} else if sum != recorded {
				rep.Addf("supply of %q is recorded as %d but %d is held",
					denom, recorded, sum)
			}
			delete(held, denom)
		}
		if err := iter.Error(); err != nil {
			rep.Addf("iteration over %q failed, so the sweep is incomplete: %v", SupplyPrefix, err)
		}
		// Whatever is left was never recorded: value exists that nothing minted.
		// Sorted, because the report is truncated: ranging the map directly made two
		// operators inspecting identical state see different findings.
		for _, denom := range slices.Sorted(maps.Keys(held)) {
			rep.Addf("%d%s is held but there is no supply record for it", held[denom], denom)
		}
	})
}

// AllInvariants runs every bank invariant.
func AllInvariants(view ViewKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []sdk.Invariant{
			BalanceKeysInvariant(view),
			AccountTierInvariant(view),
			SupplyInvariant(view),
		} {
			if msg, broken := inv(ctx); broken {
				return msg, true
			}
		}
		return "", false
	}
}
