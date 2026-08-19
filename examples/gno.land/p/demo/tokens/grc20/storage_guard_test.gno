package grc20

import (
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// These guards are exercised with avl only, so they cost grc20's on-chain
// package closure nothing — avl is already a production import. Coverage of
// the other backends lives in filetests/, which never ships on chain.

// mustPanicWith asserts that fn panics with want. newTestToken wraps NewToken
// in cross(...), so the panic crosses a realm boundary — use revive(), as
// defer-recover does not see cross-realm panics.
func mustPanicWith(t *testing.T, want error, label string, fn func()) {
	t.Helper()
	r := revive(fn)
	if r == nil {
		t.Errorf("%s: expected panic %v, got none", label, want)
		return
	}
	if r != want {
		t.Errorf("%s: expected %v, got %v", label, want, r)
	}
}

// WithStorage panics in the caller's own frame, so this one is a plain
// defer-recover rather than revive().
func TestWithStorageRejectsNilConstructor(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("nil constructor: expected panic %v, got none", ErrNilStorage)
			return
		}
		if r != ErrNilStorage {
			t.Errorf("nil constructor: expected %v, got %v", ErrNilStorage, r)
		}
	}()
	WithStorage(nil)
}

func TestNewTokenRejectsNilStore(cur realm, t *testing.T) {
	mustPanicWith(t, ErrNilStorage, "constructor returning nil", func() {
		newTestToken("Nil", "NIL", 4, 0, cur,
			WithStorage(func() KV { return nil }))
	})
}

// One shared store for both maps would silently make KnownAccounts count
// allowance rows as holders.
func TestNewTokenRejectsSharedStore(cur realm, t *testing.T) {
	shared := avl.NewTree()
	mustPanicWith(t, ErrSharedStorage, "shared store", func() {
		newTestToken("Shared", "SH", 4, 0, cur,
			WithStorage(func() KV { return shared }))
	})
}

// A pre-populated store would yield balances that totalSupply never accounted
// for — a token reporting a holder balance with zero supply.
func TestNewTokenRejectsNonEmptyStore(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	mustPanicWith(t, ErrNonEmptyStorage, "pre-populated store", func() {
		newTestToken("Seeded", "SEED", 4, 0, cur,
			WithStorage(func() KV {
				seeded := avl.NewTree()
				seeded.Set(string(alice), int64(999999))
				return seeded
			}))
	})
}

// The guards must not reject a well-behaved constructor.
func TestNewTokenAcceptsFreshStores(cur realm, t *testing.T) {
	tok, led := newTestToken("Fresh", "FRESH", 4, 0, cur,
		WithStorage(func() KV { return avl.NewTree() }))
	urequire.True(t, tok != nil, "token should be created")

	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	urequire.NoError(t, led.Mint(alice, 1000))
	urequire.NoError(t, led.Approve(alice, bob, 250))

	// balances and allowances stay in separate stores: one holder, not two.
	uassert.Equal(t, 1, tok.KnownAccounts())
	uassert.Equal(t, int64(1000), tok.BalanceOf(alice))
	uassert.Equal(t, int64(250), tok.Allowance(alice, bob))
	uassert.Equal(t, int64(1000), tok.TotalSupply())
}

// mapKV is a KV whose dynamic type is a struct value holding a map, so two
// such interface values are uncomparable. The shared-store check must not turn
// that into a runtime panic — an uncomparable pair is distinct by definition.
type mapKV struct{ m map[string]any }

func (k mapKV) Get(key string) any            { return k.m[key] }
func (k mapKV) Has(key string) bool           { _, ok := k.m[key]; return ok }
func (k mapKV) Size() int                     { return len(k.m) }
func (k mapKV) Set(key string, v any) bool    { _, u := k.m[key]; k.m[key] = v; return u }
func (k mapKV) Remove(key string) (any, bool) { v, ok := k.m[key]; delete(k.m, key); return v, ok }

func TestNewTokenAcceptsUncomparableStores(cur realm, t *testing.T) {
	tok, led := newTestToken("Uncmp", "UNCMP", 4, 0, cur,
		WithStorage(func() KV { return mapKV{m: map[string]any{}} }))
	urequire.True(t, tok != nil, "uncomparable KV should be accepted")

	alice := testutils.TestAddress("alice")
	urequire.NoError(t, led.Mint(alice, 500))
	uassert.Equal(t, int64(500), tok.BalanceOf(alice))
	uassert.Equal(t, 1, tok.KnownAccounts())
}

// The same type, genuinely shared, must still be rejected — the check has to
// stay useful for uncomparable types, not just fail open.
func TestNewTokenRejectsSharedUncomparableStore(cur realm, t *testing.T) {
	shared := mapKV{m: map[string]any{}}
	mustPanicWith(t, ErrSharedStorage, "shared uncomparable store", func() {
		newTestToken("SharedU", "SHU", 4, 0, cur,
			WithStorage(func() KV { return shared }))
	})
}
