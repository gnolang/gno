package vm

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the per-realm chain/params storage deposit flow.
// See gno.land/pkg/sdk/vm/params_deposit.go for the implementation.

func TestParamsDepositSetBytesLocks(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("deployer"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/test/dep"
	files := []*std.MemFile{
		{Name: "dep.gno", Body: `
package dep
import params_ "chain/params"
func SetK(cur realm, val string) { params_.SetBytes("k", []byte(val)) }
func DelK(cur realm)             { params_.SetBytes("k", nil) }
`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))

	// Snapshot deposit address balance after AddPackage but before our Set.
	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	depBefore := env.bankk.GetCoins(ctx, depAddr)

	// Call SetK with a 100-byte value.
	const valLen = 100
	bigBlob := make([]byte, valLen)
	for i := range bigBlob {
		bigBlob[i] = byte(i)
	}
	call := NewMsgCall(addr, std.Coins{}, pkgPath, "SetK", []string{string(bigBlob)})
	call.MaxDeposit = std.MustParseCoins(ugnot.ValueString(10_000_000))
	_, err := env.vmk.Call(ctx, call)
	require.NoError(t, err)

	// Expected delta: value bytes + key bytes ("vm:gno.land/r/test/dep:k").
	expectedKeyLen := len("vm:" + pkgPath + ":k")
	expectedDelta := int64(valLen + expectedKeyLen)

	// Read meta-key directly via the params keeper.
	got := readMetaKey(t, env, ctx, pkgPath)
	assert.Equal(t, expectedDelta, got, "meta total should equal value+key bytes")

	// Deposit address balance should have grown by expectedDelta * StoragePrice.
	storagePrice := std.MustParseCoin(env.vmk.GetParams(ctx).StoragePrice)
	expectedLock := expectedDelta * storagePrice.Amount
	depAfter := env.bankk.GetCoins(ctx, depAddr)
	assert.Equal(t, expectedLock, depAfter.AmountOf(ugnot.Denom)-depBefore.AmountOf(ugnot.Denom),
		"deposit address should hold value+key * StoragePrice extra")
}

func TestParamsDepositUpdateRefunds(t *testing.T) {
	env, ctx, addr, pkgPath := setupParamsDepRealm(t)
	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)

	// Tx1: Set 100 bytes.
	callSet(t, env, ctx, addr, pkgPath, make([]byte, 100))
	depAfterFirst := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	metaAfterFirst := readMetaKey(t, env, ctx, pkgPath)

	// Tx2: Set 50 bytes (overwrite, same key).
	callSet(t, env, ctx, addr, pkgPath, make([]byte, 50))
	depAfterSecond := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	metaAfterSecond := readMetaKey(t, env, ctx, pkgPath)

	// Update doesn't change key bytes; only the 50-byte value diff.
	storagePrice := std.MustParseCoin(env.vmk.GetParams(ctx).StoragePrice).Amount
	assert.Equal(t, int64(50), metaAfterFirst-metaAfterSecond, "meta should drop by 50 (value-only diff)")
	assert.Equal(t, int64(50)*storagePrice, depAfterFirst-depAfterSecond,
		"deposit should refund 50 bytes' worth")
}

func TestParamsDepositDeleteRefundsFull(t *testing.T) {
	env, ctx, addr, pkgPath := setupParamsDepRealm(t)
	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)

	depBefore := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	callSet(t, env, ctx, addr, pkgPath, make([]byte, 100))
	depAfterSet := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)

	// Delete returns key+value bytes worth of deposit.
	callDel(t, env, ctx, addr, pkgPath)
	depAfterDel := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	metaAfterDel := readMetaKey(t, env, ctx, pkgPath)

	assert.Equal(t, int64(0), metaAfterDel, "meta total should be 0 after delete")
	assert.Equal(t, depBefore, depAfterDel, "deposit should return to pre-Set balance")
	assert.Greater(t, depAfterSet, depBefore, "Set should have locked deposit (sanity)")
}

func TestParamsDepositPreFeatureClampNoPanic(t *testing.T) {
	env, ctx, addr, pkgPath := setupParamsDepRealm(t)

	// Inject a "pre-feature" key directly into the params store
	// (bypassing recordParamsDelta). The realm's meta-key remains 0.
	env.prmk.SetBytes(ctx, "vm:"+pkgPath+":k", []byte("pre-feature data"))

	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	depBefore := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)

	// Realm tries to delete the pre-feature key. Without the floor/clamp,
	// processStorageDeposit would panic at keeper.go:1487 because
	// rlm.Deposit < depositUnlocked. With the clamp, delta is forced to
	// 0 and no refund is attempted.
	require.NotPanics(t, func() {
		callDel(t, env, ctx, addr, pkgPath)
	})

	depAfter := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	assert.Equal(t, depBefore, depAfter, "no deposit movement on pre-feature delete")
	assert.Equal(t, int64(0), readMetaKey(t, env, ctx, pkgPath), "meta stays at 0 (floored)")
}

// Mixed pre/post-feature delete in create-then-delete order: realm has
// pre-feature data on disk, then post-feature SetK builds up meta and
// deposit, then deletes the pre-feature key. The floor doesn't fire
// (a.bytes stays positive because post-feature meta > pre-feature
// delete), so delta flows through as a refund. No panic — refund stays
// within rlm.Deposit — but the realm receives credit for bytes it
// never locked. See params_deposit.go's "Floor/clamp" comment for the
// full bounded-leak rationale; this test pins down the observable
// behavior so changes to the floor/clamp logic surface here.
func TestParamsDepositMixedPreFeatureCreateThenDeleteLeaks(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("deployer"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/test/mixedpre"
	// Realm exposes BOTH SetK (post-feature 200-byte set) and
	// DelPreX (delete the pre-injected pre-feature key).
	files := []*std.MemFile{
		{Name: "dep.gno", Body: `
package dep
import params_ "chain/params"
func SetK(cur realm, val string)  { params_.SetBytes("k", []byte(val)) }
func DelPreX(cur realm)           { params_.SetBytes("preX", nil) }
`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))

	// Inject a 100-byte pre-feature value at "preX" directly into the
	// store, bypassing recordParamsDelta (so no meta, no deposit lock).
	env.prmk.SetBytes(ctx, "vm:"+pkgPath+":preX", []byte(strings.Repeat("p", 100)))

	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	depBefore := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)

	// TX1: SetK with 200-byte value — locks (keyLen+200)*price.
	callSet(t, env, ctx, addr, pkgPath, make([]byte, 200))
	metaAfterSet := readMetaKey(t, env, ctx, pkgPath)
	depAfterSet := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	t.Logf("after SetK: meta=%d deposit=%d", metaAfterSet, depAfterSet-depBefore)

	// TX2: DelPreX. d = -(keyLen("vm:<path>:preX") + 100). Meta loaded
	// is metaAfterSet (~keyLen+200). |d| < meta → floor doesn't fire,
	// delta flows through as a refund.
	c := NewMsgCall(addr, std.Coins{}, pkgPath, "DelPreX", nil)
	c.MaxDeposit = std.MustParseCoins(ugnot.ValueString(10_000_000))
	require.NotPanics(t, func() {
		_, err := env.vmk.Call(ctx, c)
		require.NoError(t, err)
	})
	depAfterDel := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	metaAfterDel := readMetaKey(t, env, ctx, pkgPath)

	// Concrete invariants that pin the leak:
	preXKeyLen := int64(len("vm:" + pkgPath + ":preX"))
	expectedRefundedBytes := preXKeyLen + 100
	storagePrice := std.MustParseCoin(env.vmk.GetParams(ctx).StoragePrice).Amount
	assert.Equal(t, metaAfterSet-expectedRefundedBytes, metaAfterDel,
		"meta drops by full key+val of the pre-feature delete")
	assert.Equal(t, expectedRefundedBytes*storagePrice,
		depAfterSet-depAfterDel,
		"deposit refunded for pre-feature bytes — the bounded leak documented in params_deposit.go")
	// The realm now stores 229 bytes (from SetK) but its deposit only
	// covers (229 - expectedRefundedBytes) = ~96 bytes worth — short
	// by ~133 bytes' worth of locked deposit relative to actual on-disk.
	t.Logf("leak summary: on-disk=%d bytes, deposit-covered=%d bytes (short by %d bytes)",
		metaAfterSet, metaAfterDel, expectedRefundedBytes)
}

func TestParamsDepositInsufficientFails(t *testing.T) {
	env, ctx, addr, pkgPath := setupParamsDepRealm(t)
	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)

	// Snapshot balance after AddPackage (which locks deposit for the
	// realm's own code bytes). The failing call should not move it.
	depBefore := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	metaBefore := readMetaKey(t, env, ctx, pkgPath)

	// Set a large value with insufficient MaxDeposit.
	bigBlob := make([]byte, 10_000)
	call := NewMsgCall(addr, std.Coins{}, pkgPath, "SetK", []string{string(bigBlob)})
	call.MaxDeposit = std.MustParseCoins(ugnot.ValueString(100)) // 100 ugnot, not enough
	_, err := env.vmk.Call(ctx, call)
	require.Error(t, err, "should fail on insufficient deposit")

	// Meta-key unchanged — flush only fires after lock success, and
	// failed message rolls back via cache.Store discard.
	assert.Equal(t, metaBefore, readMetaKey(t, env, ctx, pkgPath))
	assert.Equal(t, depBefore, env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom))
}

// TestParamsRecordDeltaSkipsNonRealmKeys is a focused unit test for
// recordParamsDelta's prefix filter — sys/params keys (any key without
// a "vm:" prefix) must not accumulate into any realm bucket.
func TestParamsRecordDeltaSkipsNonRealmKeys(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	ctx = ContextWithParamsAccum(ctx)

	// Various non-vm key shapes plus sys/params writes that DO start
	// with "vm:" but address a bare submodule (e.g. from
	// sys/params.SetSysParamString("vm","bar","baz",...)). All must be
	// skipped — only "vm:<rlmPath>:<key>" with rlmPath containing "/"
	// is realm-attributable.
	for _, k := range []string{
		"mod:sub:name",                // sys/params shape (no "vm:" prefix)
		realmMetaPrefix + "any/realm", // meta-key under reserved prefix
		"plain_key",                   // no colon → realmFromKey rejects
		"vm:bar:baz",                  // sys/params via "vm" module — bare submodule, no "/"
		"vm:p:something",              // VM-internal config writes ("vm:p" struct)
	} {
		recordParamsDelta(ctx, env.prmk, k, 100)
	}

	diffs := ParamsRealmDiffs(ctx)
	assert.Empty(t, diffs, "no realm should have accumulated a delta from non-vm keys")
}

func TestParamsDepositMultiRealmIndependent(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("deployer"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, initialBalance)

	pkg1 := "gno.land/r/test/multi1"
	pkg2 := "gno.land/r/test/multi2"
	for _, p := range []string{pkg1, pkg2} {
		files := []*std.MemFile{
			{Name: "dep.gno", Body: `
package dep
import params_ "chain/params"
func SetK(cur realm, val string) { params_.SetBytes("k", []byte(val)) }
`},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(p)},
		}
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, p, files)))
	}

	dep1 := gnolang.DeriveStorageDepositCryptoAddr(pkg1)
	dep2 := gnolang.DeriveStorageDepositCryptoAddr(pkg2)
	dep1Before := env.bankk.GetCoins(ctx, dep1).AmountOf(ugnot.Denom)
	dep2Before := env.bankk.GetCoins(ctx, dep2).AmountOf(ugnot.Denom)

	callSet(t, env, ctx, addr, pkg1, make([]byte, 100))
	callSet(t, env, ctx, addr, pkg2, make([]byte, 200))

	storagePrice := std.MustParseCoin(env.vmk.GetParams(ctx).StoragePrice).Amount
	dep1Delta := env.bankk.GetCoins(ctx, dep1).AmountOf(ugnot.Denom) - dep1Before
	dep2Delta := env.bankk.GetCoins(ctx, dep2).AmountOf(ugnot.Denom) - dep2Before

	// realm1: 100 bytes value + 26-byte key ("vm:gno.land/r/test/multi1:k") = 126 bytes
	// realm2: 200 bytes value + 26-byte key                                  = 226 bytes
	expected1 := int64(100+len("vm:"+pkg1+":k")) * storagePrice
	expected2 := int64(200+len("vm:"+pkg2+":k")) * storagePrice
	assert.Equal(t, expected1, dep1Delta, "realm1 deposit")
	assert.Equal(t, expected2, dep2Delta, "realm2 deposit")

	assert.Equal(t, int64(100+len("vm:"+pkg1+":k")), readMetaKey(t, env, ctx, pkg1))
	assert.Equal(t, int64(200+len("vm:"+pkg2+":k")), readMetaKey(t, env, ctx, pkg2))
}

// TestParamsDepositEmptyValueIsCreate exercises the bug case the
// keysize refactor was written to fix: SetBytes(k, []byte{}) on an
// absent key is a *create* (key bytes lock), while SetBytes(k, nil) on
// the same absent key is a no-op (no movement). The pre-refactor logic
// conflated the two via `newSize == 0`. Subtests run in order; each
// observes state left by the previous one.
func TestParamsDepositEmptyValueIsCreate(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("deployer"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/test/empty"
	files := []*std.MemFile{
		{Name: "dep.gno", Body: `
package dep
import params_ "chain/params"
func SetEmpty(cur realm) { params_.SetBytes("k", []byte{}) }
func SetNil(cur realm)   { params_.SetBytes("k", nil) }
`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))

	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	storagePrice := std.MustParseCoin(env.vmk.GetParams(ctx).StoragePrice).Amount
	keyLen := int64(len("vm:" + pkgPath + ":k"))

	call := func(t *testing.T, fn string) {
		t.Helper()
		c := NewMsgCall(addr, std.Coins{}, pkgPath, fn, nil)
		c.MaxDeposit = std.MustParseCoins(ugnot.ValueString(10_000_000))
		_, err := env.vmk.Call(ctx, c)
		require.NoError(t, err)
	}

	// Snapshot once; subtests reference it as the baseline.
	depBefore := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)

	t.Run("SetNil_on_absent_is_noop", func(t *testing.T) {
		call(t, "SetNil")
		assert.Equal(t, depBefore, env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom),
			"SetBytes(nil) on absent key must not move deposit")
		assert.Equal(t, int64(0), readMetaKey(t, env, ctx, pkgPath),
			"meta stays 0 after no-op delete")
	})

	t.Run("SetEmpty_on_absent_locks_key_bytes", func(t *testing.T) {
		call(t, "SetEmpty")
		assert.Equal(t, keyLen, readMetaKey(t, env, ctx, pkgPath),
			"empty value still creates entry; meta == key bytes")
		assert.Equal(t, keyLen*storagePrice,
			env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)-depBefore,
			"empty-value create locks key bytes worth of deposit")
	})

	t.Run("SetNil_on_empty_refunds_key_bytes", func(t *testing.T) {
		call(t, "SetNil")
		assert.Equal(t, int64(0), readMetaKey(t, env, ctx, pkgPath),
			"meta returns to 0 after delete")
		assert.Equal(t, depBefore, env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom),
			"delete refunds the key bytes")
	})
}

// ---- helpers ----

// setupParamsDepRealm provisions an addr with funds and deploys a small
// realm exposing SetK/DelK. Returns env, the SHARED ctx (must reuse
// across calls so the gno transaction store stays warm), the deployer
// addr, and the realm path.
func setupParamsDepRealm(t *testing.T) (testEnv, sdk.Context, crypto.Address, string) {
	t.Helper()
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	addr := crypto.AddressFromPreimage([]byte("deployer"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, initialBalance)
	const pkgPath = "gno.land/r/test/dep"
	files := []*std.MemFile{
		{Name: "dep.gno", Body: `
package dep
import params_ "chain/params"
func SetK(cur realm, val string) { params_.SetBytes("k", []byte(val)) }
func DelK(cur realm)             { params_.SetBytes("k", nil) }
`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))
	return env, ctx, addr, pkgPath
}

func callSet(t *testing.T, env testEnv, ctx sdk.Context, addr crypto.Address, pkgPath string, val []byte) {
	t.Helper()
	c := NewMsgCall(addr, std.Coins{}, pkgPath, "SetK", []string{string(val)})
	c.MaxDeposit = std.MustParseCoins(ugnot.ValueString(10_000_000))
	_, err := env.vmk.Call(ctx, c)
	require.NoError(t, err)
}

func callDel(t *testing.T, env testEnv, ctx sdk.Context, addr crypto.Address, pkgPath string) {
	t.Helper()
	c := NewMsgCall(addr, std.Coins{}, pkgPath, "DelK", nil)
	c.MaxDeposit = std.MustParseCoins(ugnot.ValueString(10_000_000))
	_, err := env.vmk.Call(ctx, c)
	require.NoError(t, err)
}

// readMetaKey reads the persisted byte total stored under the
// per-realm meta-key. Returns 0 if the key doesn't exist.
func readMetaKey(t *testing.T, env testEnv, ctx sdk.Context, pkgPath string) int64 {
	t.Helper()
	var bz []byte
	if !env.prmk.GetBytes(ctx, realmMetaPrefix+pkgPath, &bz) {
		return 0
	}
	if len(bz) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}
