package vm

import (
	"encoding/binary"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// Per-realm chain/params byte accounting. Lives in gno.land/vm so realm
// awareness stays out of tm2. Hooks in via SDKParams (the existing
// adapter implementing execctx.ParamsInterface for the VM).
//
// Lifecycle: ContextWithParamsAccum(ctx) seeds an empty per-realm map at
// each message entry point (AddPackage / Call / Run) BEFORE the
// stdlibs.ExecContext literal that captures ctx into NewSDKParams.
// During execution, every SDKParams.Set* records a byte delta. At
// message end, processStorageDeposit merges these deltas into the
// per-realm storage diff stream and calls FlushParamsRealmAccum AFTER
// the deposit lock/refund succeeds, keeping bank state and the
// persistent meta-key consistent on partial failure.
//
// Scope:
//   - User realm writes (chain/params via SDKParams) ARE tracked.
//   - VM-internal config (vm/params.go's vm.prmk.SetStruct("vm:p", ...))
//     bypasses SDKParams. realmFromKey("vm:p") returns ("", false) so
//     any accidental routing through here is safely skipped.
//   - sys/params (governance-gated, no "vm:" prefix) is skipped.
//
// Floor/clamp asymmetry on negative bytes: at feature rollout, realms
// have no meta-key baseline (loads as 0). A delete of a pre-feature
// value would compute a.bytes < 0; we floor to 0. We then clamp delta
// to 0 to avoid asking for a refund that was never deposited (which
// would panic processStorageDeposit at keeper.go:1487). Trade-off:
// realms with mixed pre-/post-feature state may permanently lock some
// deposit. Acceptable: locked deposit > minted deposit out of thin air.

const (
	// pkey() in gnovm/stdlibs/chain/params produces "vm:<rlmPath>:<key>"
	// for user-realm writes. sys/params produces "<module>:<sub>:<name>"
	// (no "vm:" prefix). Only "vm:" is realm-attributable.
	userRealmKeyPrefix = "vm:"

	// Per-realm meta-key carrying the running byte total (8-byte int64).
	// Underscore (not colon) separator so ParamsKeeper.validate() sees
	// no module prefix and skips registered-module enforcement
	// (`tm2/pkg/sdk/params/keeper.go:261-271`). Lives outside "vm:" so
	// it does not recursively account into itself either.
	realmMetaPrefix = "_realmmeta_"
)

// realmAccum is per-message, per-realm. State lives on sdk.Context, not
// on the (singleton) SDKParams adapter or ParamsKeeper.
type realmAccum struct {
	bytes  int64 // running total: loaded baseline + tx-scoped delta
	delta  int64 // tx-scoped delta only — for processStorageDeposit
	loaded bool  // baseline already pulled from store?
	dirty  bool  // any change to flush at message end?
}

type paramsAccumCtxKey struct{}

// ContextWithParamsAccum seeds an empty per-realm accumulator. Call
// once per message, BEFORE NewSDKParams captures ctx into its struct
// field — otherwise the SDKParams holds a ctx without the accumulator
// and recordParamsDelta becomes a silent no-op.
func ContextWithParamsAccum(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(paramsAccumCtxKey{}, map[string]*realmAccum{})
}

func paramsAccum(ctx sdk.Context) map[string]*realmAccum {
	v, _ := ctx.Value(paramsAccumCtxKey{}).(map[string]*realmAccum)
	return v
}

// ParamsRealmDiffs returns the per-realm byte deltas accumulated this
// message. Filtered to non-zero entries — consumed by
// VMKeeper.processStorageDeposit to merge into the per-realm diff map.
func ParamsRealmDiffs(ctx sdk.Context) map[string]int64 {
	accum := paramsAccum(ctx)
	if accum == nil {
		return nil
	}
	out := make(map[string]int64, len(accum))
	for rlm, a := range accum {
		if a.delta != 0 {
			out[rlm] = a.delta
		}
	}
	return out
}

// FlushParamsRealmAccum writes the meta-key for one realm if dirty.
// Called by VMKeeper inside the sorted-realm loop in
// processStorageDeposit, AFTER the deposit lock/refund has succeeded —
// keeps bank state and the persistent meta-key consistent on partial
// failure (a failing lockStorageDeposit aborts the message; the
// uncommitted meta-key write would otherwise persist in cache.Store
// only to be discarded with the rest of the message).
func FlushParamsRealmAccum(ctx sdk.Context, pmk ParamsKeeperI, rlmPath string) {
	accum := paramsAccum(ctx)
	if accum == nil {
		return
	}
	a, ok := accum[rlmPath]
	if !ok || !a.dirty {
		return
	}
	pmk.SetBytes(ctx, realmMetaPrefix+rlmPath, packMeta(a.bytes))
}

// priorSize returns len(key)+len(prior value) when `key` exists, 0
// otherwise — the oldSize that recordParamsDelta expects (key bytes
// counted alongside value bytes). pmk.GetBytes does a raw stor.Get,
// which works for any Set type because the keeper stores either raw
// bytes (SetBytes) or amino-JSON (set()) under the same byte stream.
func priorSize(ctx sdk.Context, pmk ParamsKeeperI, key string) int {
	var bz []byte
	if pmk.GetBytes(ctx, key, &bz) {
		return len(key) + len(bz)
	}
	return 0
}

// Encoded sizes after a Set call — match what actually persists via the
// keeper, computed without re-reading the store. SetBytes stores raw,
// so its post-size is just len(value) inline at the caller.
func sizeAfterSetString(v string) int    { return len(amino.MustMarshalJSON(v)) }
func sizeAfterSetBool(v bool) int        { return len(amino.MustMarshalJSON(v)) }
func sizeAfterSetInt64(v int64) int      { return len(amino.MustMarshalJSON(v)) }
func sizeAfterSetUint64(v uint64) int    { return len(amino.MustMarshalJSON(v)) }
func sizeAfterSetStrings(v []string) int { return len(amino.MustMarshalJSON(v)) }

// recordParamsDelta is called from each SDKParams.Set* AFTER the keeper
// write succeeds. Skips sys/params keys (no "vm:" prefix). Lazily loads
// the persistent baseline once per realm per message.
//
// oldSize and newSize already include key bytes alongside value bytes
// (both persist on disk). Conventions enforced by callers:
//   - oldSize: 0 if absent before this write, else len(key)+len(prior value)
//   - newSize: 0 if this write deletes (SetBytes nil), else
//     len(key)+len(new value)
//
// The delta is unconditionally newSize-oldSize; create/delete/update
// fall out naturally.
func recordParamsDelta(ctx sdk.Context, pmk ParamsKeeperI, key string, oldSize, newSize int) {
	rlm, ok := realmFromKey(key)
	if !ok {
		return // sys/params or meta key
	}
	accum := paramsAccum(ctx)
	if accum == nil {
		return // ctx not seeded (e.g. some tests)
	}
	a, ok := accum[rlm]
	if !ok {
		a = &realmAccum{}
		accum[rlm] = a
	}
	if !a.loaded {
		var bz []byte
		if pmk.GetBytes(ctx, realmMetaPrefix+rlm, &bz) {
			a.bytes = unpackMeta(bz)
		}
		a.loaded = true
	}
	d := int64(newSize - oldSize)
	a.bytes += d
	floored := false
	if a.bytes < 0 {
		a.bytes = 0
		floored = true
	}
	a.delta += d
	if floored && a.delta < 0 {
		a.delta = 0 // never refund a deposit that was never locked
	}
	a.dirty = true
}

// realmFromKey parses "vm:<rlmPath>:<key>" → rlmPath. Returns ok=false
// for any non-"vm:" key, or for "vm:" keys whose middle segment isn't a
// realm path. sys/params writes (e.g. "vm:bar:baz" from
// sys/params.SetSysParamString("vm","bar","baz",...)) collide with the
// "vm:" prefix but use bare submodule names — real realm paths always
// contain "/" (e.g. "gno.land/r/foo"), so the slash check disambiguates.
func realmFromKey(key string) (string, bool) {
	if !strings.HasPrefix(key, userRealmKeyPrefix) {
		return "", false
	}
	rest := key[len(userRealmKeyPrefix):]
	colon := strings.LastIndex(rest, ":")
	if colon < 0 {
		return "", false
	}
	rlm := rest[:colon]
	if !strings.Contains(rlm, "/") {
		return "", false
	}
	return rlm, true
}

func packMeta(bytes int64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(bytes))
	return b[:]
}

func unpackMeta(b []byte) int64 {
	if len(b) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(b))
}
