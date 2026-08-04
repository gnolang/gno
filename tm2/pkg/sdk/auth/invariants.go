package auth

import (
	"bytes"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// SessionStoreKeyLen is the exact byte length of a session account key,
// /a/<master>/s/<session>.
const SessionStoreKeyLen = AccountStoreKeyLen + len(SessionStoreKeyInfix) + crypto.AddressSize

// AccountKeyKind classifies a key under AddressStoreKeyPrefix.
type AccountKeyKind int

const (
	// AccountKeyRegular is /a/<addr>, what IterateAccounts enumerates.
	AccountKeyRegular AccountKeyKind = iota
	// AccountKeySession is /a/<master>/s/<session>.
	AccountKeySession
	// AccountKeyUnknown is neither, so no iterator enumerates it. Such a key is
	// invisible to IterateAccounts, which filters on length.
	AccountKeyUnknown
)

// ParseAccountKey classifies a key under AddressStoreKeyPrefix. master is zero
// for a regular key.
//
// Length alone is not enough to classify: a key can be exactly session length
// without having the /s/ infix, and a length-only classifier would then hand it
// to a session decoder.
func ParseAccountKey(key []byte) (kind AccountKeyKind, master, addr crypto.Address) {
	if !bytes.HasPrefix(key, []byte(AddressStoreKeyPrefix)) {
		return AccountKeyUnknown, master, addr
	}
	switch len(key) {
	case AccountStoreKeyLen:
		copy(addr[:], key[len(AddressStoreKeyPrefix):])
		return AccountKeyRegular, master, addr
	case SessionStoreKeyLen:
		if !bytes.Equal(key[AccountStoreKeyLen:AccountStoreKeyLen+len(SessionStoreKeyInfix)],
			[]byte(SessionStoreKeyInfix)) {
			return AccountKeyUnknown, master, addr
		}
		copy(master[:], key[len(AddressStoreKeyPrefix):AccountStoreKeyLen])
		copy(addr[:], key[AccountStoreKeyLen+len(SessionStoreKeyInfix):])
		return AccountKeySession, master, addr
	}
	return AccountKeyUnknown, master, addr
}

// DecodeAccountSafe decodes stored account bytes where decodeAccount would panic.
//
// A zero-length value is accepted by the store and unmarshals to a nil interface
// with no error, so "err == nil" is not enough to conclude the account is usable —
// every caller must treat a nil account as a finding, not as an empty one.
func DecodeAccountSafe(bz []byte) (acc std.Account, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			acc, err = nil, fmt.Errorf("panic decoding account: %v", rec)
		}
	}()
	if err := amino.Unmarshal(bz, &acc); err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("decoded to a nil account (value was %d bytes)", len(bz))
	}
	return acc, nil
}

// HasAccount reports whether addr has an account object, without decoding it.
//
// Reads with a nil gas context: an invariant must not perturb metering. Callers
// outside this package need this because the account store key is unexported, and
// probing /a/ through another module's store handle would silently read the wrong
// store if the two were ever mounted separately.
func (ak AccountKeeper) HasAccount(ctx sdk.Context, addr crypto.Address) bool {
	return ctx.Store(ak.key).Has(nil, AddressStoreKey(addr))
}

// AccountEntry is one key under AddressStoreKeyPrefix, decoded where possible.
type AccountEntry struct {
	Key       []byte
	Kind      AccountKeyKind
	Addr      crypto.Address // the session address for a session key
	Master    crypto.Address // sessions only
	Account   std.Account    // nil iff DecodeErr != nil
	DecodeErr error
}

// IterateAccountEntries walks every key under AddressStoreKeyPrefix — regular and
// session alike, unfiltered — surfacing decode failures instead of panicking.
//
// For invariants and offline tooling. Ordinary code wants IterateAccounts, which
// filters to regular accounts and panics on a value it cannot decode. Iteration is
// gas-free; see HasAccount.
//
// Returns an error if the iteration itself failed. A store iterator can fail
// mid-sweep — the bptree store panics from Valid()/Close() if the error is never
// read, and the iavl store reports nothing at all, which reads as silent
// truncation. Callers must therefore report a non-nil error rather than concluding
// the keyspace is healthy.
func (ak AccountKeeper) IterateAccountEntries(ctx sdk.Context, cb func(AccountEntry) bool) error {
	iter := store.PrefixIterator(nil, ctx.Store(ak.key), []byte(AddressStoreKeyPrefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		e := AccountEntry{Key: append([]byte(nil), iter.Key()...)}
		e.Kind, e.Master, e.Addr = ParseAccountKey(e.Key)
		// Classify before decoding: decoding first would report a wrong-shaped key
		// as undecodable, which names the wrong problem.
		if e.Kind != AccountKeyUnknown {
			e.Account, e.DecodeErr = DecodeAccountSafe(iter.Value())
		}
		if cb(e) {
			break
		}
	}
	return iter.Error()
}

// maxUniquenessBits bounds the account-number bitset. The counter it would
// otherwise be sized from is reachable through NewAccountWithUncheckedNumber, so a
// corrupt or hostile value could ask for an allocation large enough to abort the
// process — and an out-of-memory abort is fatal, not recoverable, so no guard would
// catch it. Above this bound uniqueness is skipped and the report says so.
const maxUniquenessBits = 1 << 26 // 8 MiB of bitset, 67M accounts

// AccountKeyspaceInvariant sweeps the account keyspace.
//
// Checks that every key is a shape some iterator enumerates; that every value
// decodes to a non-nil account; that the account's own Address agrees with the key
// it is filed under; that account numbers are unique and below the global counter;
// and that every session names a master that exists and claims that same master.
//
// The key/field agreement check is the important one. SetAccount files an account
// under AddressStoreKey(acc.GetAddress()) while every reader looks it up by the
// address it already has, so a disagreement silently redirects writes: crediting
// one address can land the coins in another.
func AccountKeyspaceInvariant(ak AccountKeeper) sdk.Invariant {
	return sdk.Guard(ModuleName, "account-keyspace", func(ctx sdk.Context, rep *sdk.InvariantReport) {
		global, err := peekGlobalAccountNumber(ctx, ak)
		if err != nil {
			rep.Addf("global account number is unreadable, so number checks were "+
				"skipped: %v", err)
		}

		var seen []uint64
		checkUnique := global > 0 && global <= maxUniquenessBits
		if checkUnique {
			seen = make([]uint64, (global+63)/64)
		} else if global > maxUniquenessBits {
			rep.Addf("global account number is %d, above the %d bound this check will "+
				"allocate for; uniqueness was NOT verified", global, maxUniquenessBits)
		}

		err = ak.IterateAccountEntries(ctx, func(e AccountEntry) bool {
			switch e.Kind {
			case AccountKeyUnknown:
				rep.Addf("key %X under %q has an unrecognised shape, so no iterator "+
					"enumerates it", e.Key, AddressStoreKeyPrefix)
				return false
			case AccountKeySession:
				if e.DecodeErr != nil {
					rep.Addf("session %s of master %s does not decode: %v", e.Addr, e.Master, e.DecodeErr)
					return false
				}
				if !ak.HasAccount(ctx, e.Master) {
					rep.Addf("session %s is filed under master %s, which has no account object",
						e.Addr, e.Master)
				}
				if da, ok := e.Account.(std.DelegatedAccount); ok {
					if got := da.GetMasterAddress(); got != e.Master {
						rep.Addf("session %s is filed under master %s but claims master %s",
							e.Addr, e.Master, got)
					}
				} else {
					rep.Addf("key %X is at a session path but does not hold a delegated account", e.Key)
				}
			case AccountKeyRegular:
				if e.DecodeErr != nil {
					rep.Addf("account %s does not decode: %v", e.Addr, e.DecodeErr)
					return false
				}
				if _, ok := e.Account.(std.DelegatedAccount); ok {
					rep.Addf("account %s is a delegated (session) account filed at a "+
						"regular path, where it is enumerated as an ordinary account", e.Addr)
				}
			}

			if e.Account == nil {
				return false
			}
			// Key/field agreement: SetAccount keys off this field.
			if got := e.Account.GetAddress(); got != e.Addr {
				rep.Addf("key %X holds an account whose address is %s: a write to %s "+
					"would be filed under %s instead", e.Key, got, e.Addr, got)
			}
			// Checked unconditionally, including when the counter reads zero: every
			// number comes from the counter, which is bumped past it, so an account
			// numbered at or above it means the two disagree either way.
			num := e.Account.GetAccountNumber()
			if num >= global {
				rep.Addf("account %s has number %d, at or above the global counter %d",
					e.Addr, num, global)
				return false
			}
			if checkUnique {
				if seen[num/64]&(1<<(num%64)) != 0 {
					rep.Addf("account number %d is used more than once, again by %s", num, e.Addr)
				}
				seen[num/64] |= 1 << (num % 64)
			}
			return false
		})
		if err != nil {
			rep.Addf("iteration over %q failed, so the sweep is incomplete: %v",
				AddressStoreKeyPrefix, err)
		}
	})
}

// AllInvariants runs every auth invariant.
func AllInvariants(ak AccountKeeper) sdk.Invariant {
	return AccountKeyspaceInvariant(ak)
}

// peekGlobalAccountNumber reads the account-number counter without incrementing it
// and without panicking. GetNextAccountNumber bumps the counter, which an invariant
// must never do, and panics on a value it cannot decode, which is a state worth
// reporting rather than dying on.
func peekGlobalAccountNumber(ctx sdk.Context, ak AccountKeeper) (uint64, error) {
	bz := ctx.Store(ak.key).Get(nil, []byte(GlobalAccountNumberKey))
	if bz == nil {
		return 0, nil // never allocated
	}
	var n uint64
	if err := amino.Unmarshal(bz, &n); err != nil {
		return 0, err
	}
	return n, nil
}
