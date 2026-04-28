// Package allowancebanker provides a bounded, granter-revocable
// spending capability that wraps a banker.Banker source.
//
// # Use case
//
// Realm A wants to grant realm B the ability to spend up to some
// bounded amount from A's address (or A's tx envelope) within a single
// call, without giving B unbounded access. After A's call to B returns,
// A revokes the capability via Close(); B cannot use it across tx
// boundaries even if it persisted the reference.
//
// # Canonical pattern
//
//	src := banker.NewBanker(banker.BankerTypeRealmSend)
//	al  := allowancebanker.New(src, granterAddr,
//	           chain.NewCoins(chain.NewCoin("ugnot", 1_000_000)))
//	defer al.Close()
//	bRealm.DoSomething(cross, al)  // B can call al.Send up to cap
//
// On return (success or panic), `defer al.Close()` flips the closed
// flag. If B persisted al, B's stored reference is now dead — any
// later al.Send() panics with "closed".
//
// # Security model
//
//   - Pointer-based API. All callees that hold a *AllowanceBanker share
//     the same underlying state. Close() flips a bool that every reader
//     observes synchronously.
//   - Inner banker is an unexported field. Callees cannot extract it
//     to bypass the wrapper. Language-enforced.
//   - Cap is enforced per-denom via chain.Coins.AmountOf. Any denom in
//     amt that is missing from cap implicitly has capAmt=0; positive
//     spend on such a denom panics with "cap exceeded".
//   - Persistence is the kill switch's friend, not its enemy. The
//     closed flag persists with the struct, so a callee that stored
//     the reference for cross-tx use sees a closed allowance.
//   - Defer-close survives panic. If B's call panics, the tx reverts
//     entirely (so the close is irrelevant — state reverts anyway). On
//     success, defer fires before granter's function returns,
//     committing closed=true with the rest of the tx state.
//   - Cap/Spent return defensive copies. Mutating the returned
//     chain.Coins does not affect internal state.
//   - Inner banker panic on SendCoins (e.g. bank-keeper "insufficient
//     funds") rolls back the spent counter — caller's accounting stays
//     accurate even if a defer-recover swallows the panic. (Without
//     recovery, the entire tx reverts and the rollback is moot.)
//
// # Design choices
//
//   - Pointer type, not value: forces shared state across references.
//     Value-copy AllowanceBanker would have separate closed flags and
//     defeat the kill switch.
//   - One-shot Close, not graduated: simpler invariant. If you need
//     "resume later," create a fresh allowance.
//   - No destination allowlist: out of scope. If the granter wants to
//     restrict where funds go, wrap this further (e.g. an
//     AllowanceWithRecipientCheck that only forwards to specific tos).
//   - No time-based expiry: rely on Close. Block-height stamping or
//     deadline checks would require runtime cooperation; this package
//     is pure Gno.
//   - Cap is multi-denom (chain.Coins, not int64): supports
//     heterogeneous payment policies. Most callers use single-denom.
//   - Idempotent Close: safe to defer-Close even if Close was called
//     manually earlier in the function.
//
// # Limitations
//
//   - Granter retains direct access to the underlying source banker.
//     The AllowanceBanker only constrains the wrapped capability, not
//     the granter's ability to spend its own funds via other means.
//   - No cross-realm enforcement of granter identity. If a malicious
//     realm somehow constructs an AllowanceBanker with a forged
//     payer/inner combination, the source banker's own pkgAddr check
//     (banker.gno:116) would still reject mismatched froms. This
//     package relies on banker's own protection there.
//   - Re-entrancy: nested allowances work (each is independent), but
//     do NOT re-use the same AllowanceBanker pointer across grants —
//     always create a fresh one. Re-using would mix spent counters.
//   - Cap with duplicate denoms (constructed by hand, not via
//     chain.NewCoins) will trigger chain.Coins.AmountOf to panic on
//     read. Use chain.NewCoins(...) to construct cap; it deduplicates.
//   - int64 overflow on (spent + amt) is not explicitly guarded.
//     Practical concern only at amounts approaching MaxInt64; not a
//     concern for ugnot-scale operations.
//   - Persistence: granter is responsible for clearing references to
//     closed AllowanceBanker pointers from its own state if it wants
//     them garbage-collected; otherwise they linger as dead structs.
//
// # What this package does NOT solve
//
//   - "Allowance survives across N txs but not N+1 txs." Use a session
//     counter pattern in the granter realm; this package's Close is
//     binary, not deadline-based.
//   - "Fungible-token (grc20) allowances." Use the grc20 package's
//     own Approve/Allowance/TransferFrom triple. AllowanceBanker is
//     for native (chain-coin) sends only.
//
// # Events
//
// AllowanceBanker emits chain events at three points so off-chain
// observers can track allowance lifecycles. Events are emitted from
// the calling realm's package path (whichever realm holds the
// AllowanceBanker pointer when the method is called).
//
//   - "AllowanceBankerSend": on each successful Send, with attributes
//     "payer", "to", "amount", "spent_total", "remaining".
//   - "AllowanceBankerClose": on Close (only the first call; idempotent
//     re-Closes do not re-emit).
//
// The underlying inner banker also emits its own bank-keeper events
// for the actual coin movement; AllowanceBanker's events sit on top of
// those for allowance-level audit trails.
package allowancebanker

import (
	"chain"
	"chain/banker"
	"math/overflow"
)

// Event names emitted by AllowanceBanker. Use these constants when
// asserting on events from off-chain observers or test code.
const (
	EventSend  = "AllowanceBankerSend"
	EventClose = "AllowanceBankerClose"
)

// AllowanceBanker is a bounded, revocable spending capability over a
// source banker.Banker. Always pass *AllowanceBanker (pointer); value
// copies are not supported and would not share state.
type AllowanceBanker struct {
	inner  banker.Banker // unexported — callees cannot extract
	payer  address       // address debited; must match inner's source
	limit  chain.Coins   // maximum cumulative spend, per-denom
	spent  chain.Coins   // running total of spends, per-denom
	closed bool          // once true, all further Send calls panic
}

// New creates a new AllowanceBanker wrapping src with the given cap.
// payer is the address that inner.SendCoins will draw from (must match
// the realm address src was created against; the banker enforces this
// at SendCoins time).
//
// Panics if inner is nil or payer is invalid. cap may be empty (zero
// allowance — every Send panics with "cap exceeded"); cap may also
// contain multiple denoms.
func New(inner banker.Banker, payer address, limit chain.Coins) *AllowanceBanker {
	if inner == nil {
		panic("allowancebanker: inner banker is nil")
	}
	if !payer.IsValid() {
		panic("allowancebanker: payer address is invalid")
	}
	return &AllowanceBanker{
		inner: inner,
		payer: payer,
		limit: limit,
	}
}

// Send debits up to (cap - spent) per denom from payer to to via the
// inner banker. Updates spent on success. On any panic (including
// inner.SendCoins panic for bank-level reasons like insufficient
// balance), spent is rolled back — caller's accounting stays accurate
// even if a caller wraps Send in a defer-recover.
//
// Arithmetic uses math/overflow.Add64; an overflow on (spent+amount)
// panics with a distinct message rather than silently wrapping. This
// closes the int64-overflow vector where a malicious callee passes a
// MaxInt64-style amount to bypass the cap check.
//
// Negative amounts are rejected explicitly. chain.NewCoin permits
// signed amounts at construction; we don't accept them.
//
// Emits "AllowanceBankerSend" event on success.
//
// Panics:
//   - "closed" if Close has been called.
//   - "negative amount not allowed" if any c.Amount < 0.
//   - "amount overflow on spent+amt" if int64 addition would overflow.
//   - "cap exceeded for denom <denom>" if any denom in amt would push
//     spent over cap.
//   - inner.SendCoins panics propagate (typically "insufficient
//     funds" from the bank keeper); spent is rolled back first.
func (a *AllowanceBanker) Send(to address, amt chain.Coins) {
	if a.closed {
		panic("allowancebanker: closed")
	}

	// Per-denom validation. Done in a separate pass before any state
	// mutation so a partial failure doesn't leak into spent.
	for _, c := range amt {
		if c.Amount < 0 {
			panic("allowancebanker: negative amount not allowed for denom " + c.Denom)
		}
		capAmt := a.limit.AmountOf(c.Denom)
		spentAmt := a.spent.AmountOf(c.Denom)
		sum, ok := overflow.Add64(spentAmt, c.Amount)
		if !ok {
			panic("allowancebanker: amount overflow on spent+amt for denom " + c.Denom)
		}
		if sum > capAmt {
			panic("allowancebanker: cap exceeded for denom " + c.Denom)
		}
	}

	// Snapshot for rollback. If inner.SendCoins panics — even if a
	// caller's defer-recover swallows the panic — the deferred
	// rollback below restores spent before re-raising. Without this,
	// a recovered panic would leave spent inflated for a transfer
	// that didn't move funds at the bank-keeper level, allowing an
	// adversary to "burn" allowance against failed sends.
	prevSpent := a.spent
	a.spent = a.spent.Add(amt)

	defer func() {
		if r := recover(); r != nil {
			a.spent = prevSpent
			panic(r) // re-raise so caller learns of the failure
		}
	}()

	a.inner.SendCoins(a.payer, to, amt)

	// Reached only on successful inner.SendCoins. Emit the audit
	// event from the granter realm's perspective.
	chain.Emit(EventSend,
		"payer", a.payer.String(),
		"to", to.String(),
		"amount", amt.String(),
		"spent_total", a.spent.String(),
		"remaining", a.Remaining().String(),
	)
}

// Cap returns a defensive copy of the maximum cumulative spend.
// Mutating the returned slice does not affect internal state.
func (a *AllowanceBanker) Cap() chain.Coins {
	return cloneCoins(a.limit)
}

// Spent returns a defensive copy of the cumulative amount spent.
// Mutating the returned slice does not affect internal state.
func (a *AllowanceBanker) Spent() chain.Coins {
	return cloneCoins(a.spent)
}

// Remaining returns cap - spent, per denom. Denoms with non-positive
// remainder are omitted.
func (a *AllowanceBanker) Remaining() chain.Coins {
	if len(a.limit) == 0 {
		return nil
	}
	out := make(chain.Coins, 0, len(a.limit))
	for _, c := range a.limit {
		rem := c.Amount - a.spent.AmountOf(c.Denom)
		if rem > 0 {
			out = append(out, chain.NewCoin(c.Denom, rem))
		}
	}
	return out
}

// Closed reports whether Close has been called.
func (a *AllowanceBanker) Closed() bool {
	return a.closed
}

// Close terminates the allowance. Subsequent Send calls panic.
// Idempotent — calling twice is fine; the second call is a no-op and
// does not re-emit the AllowanceBankerClose event.
func (a *AllowanceBanker) Close() {
	if a.closed {
		return
	}
	a.closed = true
	chain.Emit(EventClose,
		"payer", a.payer.String(),
		"spent_total", a.spent.String(),
	)
}

// Payer returns the address that this allowance debits from.
func (a *AllowanceBanker) Payer() address {
	return a.payer
}

// cloneCoins returns a fresh chain.Coins that does not share an
// underlying array with the source. Used to defend internal state
// from caller mutation.
func cloneCoins(src chain.Coins) chain.Coins {
	if len(src) == 0 {
		return nil
	}
	out := make(chain.Coins, len(src))
	copy(out, src)
	return out
}
