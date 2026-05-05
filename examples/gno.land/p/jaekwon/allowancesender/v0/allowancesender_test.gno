package allowancesender

import (
	"chain"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// mockBanker implements banker.Banker for unit testing without
// touching real bank state. Records every SendCoins call.
//
// Since the public New constructor rejects non-canonical bankers via
// banker.IsCanonical, mockBanker cannot flow through New. Tests that
// need a controllable mock build the AllowanceSender via newWithMock,
// which uses in-package struct-literal construction. Same-package
// authorship is the trust boundary; the public API remains unforgeable
// to other packages.
type mockBanker struct {
	sends    []sendCall
	failNext bool   // if true, next SendCoins panics
	failMsg  string // panic message for failNext
}

type sendCall struct {
	from, to address
	amt      chain.Coins
}

func (m *mockBanker) GetCoins(addr address) chain.Coins { return nil }
func (m *mockBanker) SendCoins(from, to address, amt chain.Coins) {
	if m.failNext {
		m.failNext = false
		panic(m.failMsg)
	}
	m.sends = append(m.sends, sendCall{from, to, amt})
}
func (m *mockBanker) TotalCoin(denom string) int64                        { return 0 }
func (m *mockBanker) IssueCoin(addr address, denom string, amount int64)  {}
func (m *mockBanker) RemoveCoin(addr address, denom string, amount int64) {}

// newWithMock constructs an *AllowanceSender wrapping a mockBanker via
// struct-literal initialization, bypassing New (which would reject the
// mock as non-canonical). Used only in this test file.
func newWithMock(mb *mockBanker, p address, limit chain.Coins) *AllowanceSender {
	return &AllowanceSender{
		inner: mb,
		payer: p,
		limit: limit,
	}
}

var (
	payer = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	bob   = address("g1mtmrdmqfu0aryqfl4aw65n35haw2wdjkh5p4cp")
	carol = address("g16dverxg6s6vzha89k22c3zusx6sug5ycjyjq2j")
)

// ---------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------

func TestNew_Valid(t *testing.T) {
	mb := &mockBanker{}
	limit := chain.NewCoins(chain.NewCoin("ugnot", 1000))
	al := newWithMock(mb, payer, limit)

	uassert.Equal(t, false, al.Closed())
	uassert.Equal(t, payer.String(), al.Payer().String())
	uassert.Equal(t, limit.String(), al.Cap().String())
}

func TestNew_NonCanonicalBanker(t *testing.T) {
	// A hand-rolled banker.Banker (mockBanker) is not the canonical
	// chain/banker.Banker; New must reject it. This is the security
	// hole the IsCanonical gate closes — without it, a malicious caller
	// could pass a no-op banker and cause AllowanceSender.Send to
	// silently succeed without moving real coins.
	mb := &mockBanker{}
	uassert.PanicsWithMessage(t,
		"allowancesender: inner banker is not the canonical chain/banker.Banker",
		func() {
			New(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1)))
		})
}

func TestNew_NilInner(t *testing.T) {
	// Nil also fails IsCanonical (nil interface assertion is false).
	uassert.PanicsWithMessage(t,
		"allowancesender: inner banker is not the canonical chain/banker.Banker",
		func() {
			New(nil, payer, chain.NewCoins(chain.NewCoin("ugnot", 1)))
		})
}

func TestNew_EmptyCap(t *testing.T) {
	// Empty cap is valid at construction (every Send will panic with
	// "cap exceeded"). Useful for "preallocated zero-allowance" cases.
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.Coins{})
	uassert.Equal(t, false, al.Closed())
}

// ---------------------------------------------------------------------
// Send happy-path tests
// ---------------------------------------------------------------------

func TestSend_WithinCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 400)))
	})

	uassert.Equal(t, 1, len(mb.sends))
	uassert.Equal(t, payer.String(), mb.sends[0].from.String())
	uassert.Equal(t, bob.String(), mb.sends[0].to.String())
	uassert.Equal(t, int64(400), mb.sends[0].amt.AmountOf("ugnot"))
}

func TestSend_MultipleWithinCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 300)))
	})
	urequire.NotPanics(t, func() {
		al.Send(carol, chain.NewCoins(chain.NewCoin("ugnot", 600)))
	})

	uassert.Equal(t, 2, len(mb.sends))
	uassert.Equal(t, int64(900), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(100), al.Remaining().AmountOf("ugnot"))
}

func TestSend_ExactlyAtCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 1000)))
	})
	uassert.Equal(t, int64(1000), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, 0, len(al.Remaining()))
}

// ---------------------------------------------------------------------
// Send rejection tests
// ---------------------------------------------------------------------

func TestSend_ExceedsCapInOneCall(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	uassert.PanicsWithMessage(t, "allowancesender: cap exceeded for denom ugnot", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 1001)))
	})
	// State must not change on rejection.
	uassert.Equal(t, 0, len(mb.sends))
	uassert.Equal(t, int64(0), al.Spent().AmountOf("ugnot"))
}

func TestSend_ExceedsCapAcrossCalls(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 700)))
	})
	uassert.PanicsWithMessage(t, "allowancesender: cap exceeded for denom ugnot", func() {
		al.Send(carol, chain.NewCoins(chain.NewCoin("ugnot", 400)))
	})
	// First send happened, second didn't.
	uassert.Equal(t, 1, len(mb.sends))
	uassert.Equal(t, int64(700), al.Spent().AmountOf("ugnot"))
}

func TestSend_DenomNotInCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// "atom" is not in cap; capAmt=0; any positive amount panics.
	uassert.PanicsWithMessage(t, "allowancesender: cap exceeded for denom atom", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("atom", 1)))
	})
}

func TestSend_ZeroCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.Coins{}) // empty cap

	uassert.PanicsWithMessage(t, "allowancesender: cap exceeded for denom ugnot", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 1)))
	})
}

func TestSend_AfterClose(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	al.Close()
	uassert.True(t, al.Closed())

	uassert.PanicsWithMessage(t, "allowancesender: closed", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 100)))
	})
	uassert.Equal(t, 0, len(mb.sends))
}

// ---------------------------------------------------------------------
// Multi-denom tests
// ---------------------------------------------------------------------

func TestSend_MultiDenomCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(
		chain.NewCoin("ugnot", 1000),
		chain.NewCoin("foo", 50),
	))

	// Spend on both denoms within their caps.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(
			chain.NewCoin("ugnot", 600),
			chain.NewCoin("foo", 30),
		))
	})

	uassert.Equal(t, int64(600), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(30), al.Spent().AmountOf("foo"))

	rem := al.Remaining()
	uassert.Equal(t, int64(400), rem.AmountOf("ugnot"))
	uassert.Equal(t, int64(20), rem.AmountOf("foo"))
}

func TestSend_OneDenomExceedsBlocksAll(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(
		chain.NewCoin("ugnot", 1000),
		chain.NewCoin("foo", 50),
	))

	// foo would exceed (51 > 50); whole send rejected, ugnot doesn't go through.
	uassert.PanicsWithMessage(t, "allowancesender: cap exceeded for denom foo", func() {
		al.Send(bob, chain.NewCoins(
			chain.NewCoin("ugnot", 100),
			chain.NewCoin("foo", 51),
		))
	})

	// Neither denom should have been spent.
	uassert.Equal(t, int64(0), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(0), al.Spent().AmountOf("foo"))
	uassert.Equal(t, 0, len(mb.sends))
}

// ---------------------------------------------------------------------
// Close behavior
// ---------------------------------------------------------------------

func TestClose_Idempotent(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	al.Close()
	urequire.NotPanics(t, func() { al.Close() })
	urequire.NotPanics(t, func() { al.Close() })
	uassert.True(t, al.Closed())
}

func TestClose_DoesNotInterfereWithReadMethods(t *testing.T) {
	mb := &mockBanker{}
	limit := chain.NewCoins(chain.NewCoin("ugnot", 1000))
	al := newWithMock(mb, payer, limit)

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 250)))
	})
	al.Close()

	// Read methods must keep working after close.
	uassert.Equal(t, limit.String(), al.Cap().String())
	uassert.Equal(t, int64(250), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(750), al.Remaining().AmountOf("ugnot"))
	uassert.True(t, al.Closed())
}

// ---------------------------------------------------------------------
// Inner-banker panic tests
// ---------------------------------------------------------------------

func TestSend_InnerPanicRollsBackSpent(t *testing.T) {
	mb := &mockBanker{failNext: true, failMsg: "insufficient funds"}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	uassert.PanicsWithMessage(t, "insufficient funds", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 500)))
	})

	// On inner panic, spent must be rolled back.
	uassert.Equal(t, int64(0), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(1000), al.Remaining().AmountOf("ugnot"))

	// And subsequent valid Send still works.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 200)))
	})
	uassert.Equal(t, int64(200), al.Spent().AmountOf("ugnot"))
}

// ---------------------------------------------------------------------
// Remaining / Spent edge cases
// ---------------------------------------------------------------------

func TestRemaining_EmptyCap(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.Coins{})

	uassert.Equal(t, 0, len(al.Remaining()))
}

func TestRemaining_FullySpent(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 100)))

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 100)))
	})
	uassert.Equal(t, 0, len(al.Remaining()))
}

func TestRemaining_PartialAcrossDenoms(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(
		chain.NewCoin("ugnot", 1000),
		chain.NewCoin("foo", 50),
	))

	// Spend all of foo, half of ugnot.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(
			chain.NewCoin("ugnot", 500),
			chain.NewCoin("foo", 50),
		))
	})

	rem := al.Remaining()
	// foo should be omitted (zero remainder); ugnot has 500 left.
	uassert.Equal(t, 1, len(rem))
	uassert.Equal(t, int64(500), rem.AmountOf("ugnot"))
	uassert.Equal(t, int64(0), rem.AmountOf("foo"))
}

// ---------------------------------------------------------------------
// Pointer semantics: shared state across copies
// ---------------------------------------------------------------------

func TestPointerSemantics_CloseVisibleViaAlias(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// Aliasing the pointer should share state.
	alias := al
	urequire.NotPanics(t, func() {
		alias.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 100)))
	})
	uassert.Equal(t, int64(100), al.Spent().AmountOf("ugnot"))

	// Close via either reference.
	alias.Close()
	uassert.True(t, al.Closed())
	uassert.True(t, alias.Closed())

	uassert.PanicsWithMessage(t, "allowancesender: closed", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 1)))
	})
}

// ---------------------------------------------------------------------
// Defensive copy: mutating returned Coins must not leak into state
// ---------------------------------------------------------------------

func TestCap_ReturnsDefensiveCopy(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// Mutate the returned cap; internal state must not change.
	got := al.Cap()
	if len(got) > 0 {
		got[0].Amount = 9_999_999_999
	}

	// Confirm internal state intact: re-read Cap.
	again := al.Cap()
	uassert.Equal(t, int64(1000), again.AmountOf("ugnot"))

	// Confirm enforcement still works: 1001 should still be rejected.
	uassert.PanicsWithMessage(t, "allowancesender: cap exceeded for denom ugnot", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 1001)))
	})
}

func TestSpent_ReturnsDefensiveCopy(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 300)))
	})

	got := al.Spent()
	if len(got) > 0 {
		got[0].Amount = 0 // try to "rewind" spent
	}

	// Internal spent must remain accurate.
	uassert.Equal(t, int64(300), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(700), al.Remaining().AmountOf("ugnot"))
}

// ---------------------------------------------------------------------
// Independence of separate AllowanceSenders
// ---------------------------------------------------------------------

func TestMultipleAllowances_Independent(t *testing.T) {
	mb := &mockBanker{}
	a1 := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 100)))
	a2 := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 200)))

	urequire.NotPanics(t, func() {
		a1.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 80)))
	})
	urequire.NotPanics(t, func() {
		a2.Send(carol, chain.NewCoins(chain.NewCoin("ugnot", 150)))
	})

	uassert.Equal(t, int64(80), a1.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(150), a2.Spent().AmountOf("ugnot"))

	// Closing one must not close the other.
	a1.Close()
	uassert.True(t, a1.Closed())
	uassert.False(t, a2.Closed())

	urequire.NotPanics(t, func() {
		a2.Send(carol, chain.NewCoins(chain.NewCoin("ugnot", 50)))
	})
	uassert.PanicsWithMessage(t, "allowancesender: closed", func() {
		a1.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 1)))
	})
}

// ---------------------------------------------------------------------
// Overflow / negative-amount adversarial inputs
// ---------------------------------------------------------------------

// Adversary attempts to bypass the cap check by passing a value that
// overflows int64 when added to an existing spent counter. Must panic
// with the overflow-specific message rather than silently wrapping
// (which would let the cap check pass and allow an absurd send).
func TestSend_OverflowAttempt(t *testing.T) {
	mb := &mockBanker{}
	const maxInt64 = int64(9223372036854775807)
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", maxInt64)))

	// First spend uses most of the cap.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", maxInt64-100)))
	})

	// Now spent ≈ MaxInt64 - 100. An adversary tries to pass amount
	// 200, which would overflow when added to spent. Must panic with
	// "amount overflow", not silently wrap to negative and bypass cap.
	uassert.PanicsWithMessage(t, "allowancesender: amount overflow on spent+amt for denom ugnot", func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 200)))
	})

	// State must not have changed.
	uassert.Equal(t, maxInt64-100, al.Spent().AmountOf("ugnot"))
	// Bank-side recorded only the first valid send.
	uassert.Equal(t, 1, len(mb.sends))
}

func TestSend_NegativeAmount(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// Spend something so we have a non-zero spent.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 200)))
	})

	// Adversary tries to "uncharge" the allowance by passing a negative
	// amount. Without this guard, spent would decrease, allowing more
	// future spends than originally permitted.
	uassert.PanicsWithMessage(t, "allowancesender: negative amount not allowed for denom ugnot", func() {
		al.Send(bob, chain.Coins{chain.Coin{Denom: "ugnot", Amount: -50}})
	})

	// Spent is unchanged after the rejected attempt.
	uassert.Equal(t, int64(200), al.Spent().AmountOf("ugnot"))
}

// Adversary's defer-recover swallows the panic from inner.SendCoins to
// hide a failed send. Verifies that even with recovery, spent has been
// rolled back — i.e., a recovered panic doesn't burn allowance against
// a transfer that never moved funds.
func TestSend_RecoveredPanicDoesNotBurnAllowance(t *testing.T) {
	mb := &mockBanker{failNext: true, failMsg: "insufficient funds"}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// Caller deliberately recovers from the inner panic.
	func() {
		defer func() {
			_ = recover() // swallow
		}()
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 500)))
	}()

	// After the recovered panic: spent must be 0 (no funds moved →
	// allowance must not be charged).
	uassert.Equal(t, int64(0), al.Spent().AmountOf("ugnot"))
	uassert.Equal(t, int64(1000), al.Remaining().AmountOf("ugnot"))

	// And subsequent valid Send (with non-failing inner) succeeds with
	// full cap available.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.NewCoins(chain.NewCoin("ugnot", 800)))
	})
	uassert.Equal(t, int64(800), al.Spent().AmountOf("ugnot"))
}

// ---------------------------------------------------------------------
// Empty/zero-amount edge cases
// ---------------------------------------------------------------------

func TestSend_EmptyAmt(t *testing.T) {
	mb := &mockBanker{}
	al := newWithMock(mb, payer, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// Sending zero coins should not panic; just no-op forwards to inner.
	urequire.NotPanics(t, func() {
		al.Send(bob, chain.Coins{})
	})

	// inner.SendCoins was still called (the wrapper doesn't filter
	// zero sends; that's the inner banker's choice). State unchanged.
	uassert.Equal(t, int64(0), al.Spent().AmountOf("ugnot"))
}
