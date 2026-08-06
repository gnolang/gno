// autofreeze.gno — Phase B: automated validator freeze via off-chain monitors.
//
// Phase A (freeze.gno) is the human lever: GovDAO, or an admin GovDAO
// has vouched for. Phase B adds a machine lever — registered off-chain
// monitors that observe misbehaviour and freeze without waiting for a
// vote — and the on-chain invariants that make it safe to hand a hot
// key that much authority.
//
// THE MONITOR IS UNTRUSTED. Assume its key is stolen the day it is
// registered and ask what the thief can do. Four invariants bound the
// answer:
//
//  1. Power floor. Post-freeze live power must stay ≥ autoFloorPct of
//     the roster's total power. Stricter than Phase A's 2/3, because a
//     monitor freeze has no human in the loop.
//
//  2. Auto-frozen cap. At most autoCapPct of roster power may sit
//     auto-frozen at once, no matter how many monitors are registered.
//
//  3. Protected set. GovDAO can mark validators no monitor may touch.
//
//  4. Auto-expiry. Entries expire after autoFreezeExpiryBlocks and
//     anyone can sweep them back in. A monitor cannot freeze anyone
//     permanently; it can only buy governance a day to look.
//
// So the worst case for a stolen monitor key is: some validators
// holding up to autoCapPct of power drop out for up to a day, and never
// enough to stop the chain. Recovery is a GovDAO deregistration, or a
// manual unfreeze, or simply waiting.
//
// Operator-keyed throughout, for the same reason Phase A is: the
// signing key is re-resolved from valoperCache on the way back in, so
// an operator who rotates while auto-frozen is never restored under a
// retired pubkey.
//
// Still open (see the PR description): N-of-M monitor attestation
// before a freeze fires, per-monitor rate limiting, and whether
// evidence should be a structured proof rather than a free-form string.
package validators

import (
	"strconv"
	"strings"

	"chain"
	"chain/runtime"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

// autoFreezeExpiryBlocks is how long an auto-frozen entry lives before
// anyone may sweep it back into the valset. ~1 day at 1s blocks: long
// enough for governance to convene, short enough that a stolen monitor
// key cannot mothball a validator.
const autoFreezeExpiryBlocks = int64(86400)

// Power bounds on the monitor path, as percentages of roster power.
const (
	// autoFloorPct: live power must stay at or above this after a
	// monitor freeze.
	autoFloorPct = 80
	// autoCapPct: total auto-frozen power may not exceed this.
	autoCapPct = 12
)

// monitors holds registered monitor addresses (addr.String() → struct{}).
var monitors = avl.NewTree()

// protectedSet holds OPERATOR addresses monitors may not auto-freeze
// (op.String() → struct{}).
var protectedSet = avl.NewTree()

// autoFrozenSet tracks auto-frozen entries separately from manual ones,
// so expiry only ever sweeps freezes a machine made.
//
// key: OPERATOR addr.String() → autoFrozenEntry
var autoFrozenSet = avl.NewTree()

type autoFrozenEntry struct {
	Op             address
	SigningAddress address
	Power          uint64
	Reason         string
	Evidence       string
	Monitor        address
	FrozenAt       int64
	ExpiresAt      int64
}

// ---- auto-freeze ---------------------------------------------------------------

// AutoFreezeValidator freezes op on a registered monitor's say-so. All
// four invariants are checked before anything is written.
//
// The entry expires on its own after autoFreezeExpiryBlocks; no monitor
// action is needed (or possible) to release it.
func AutoFreezeValidator(cur realm, op address, reason, evidence string) {
	monitorAddr := assertMonitorCaller(0, cur)

	key := op.String()
	if frozenSet.Has(key) {
		panic("operator is already manually frozen: " + key)
	}
	if autoFrozenSet.Has(key) {
		panic("operator is already auto-frozen: " + key)
	}
	// Invariant 3: protected set.
	if protectedSet.Has(key) {
		panic("operator is protected and cannot be auto-frozen: " + key)
	}

	entry := mustValoper(op)
	effective := sysparams.GetValsetEffective()

	power, live := uint64(0), false
	livePower := uint64(0)
	for _, v := range effective {
		livePower += v.VotingPower
		if v.Address == entry.SigningAddress {
			power, live = v.VotingPower, true
		}
	}
	if !live {
		panic("operator is not in the effective valset: " + key)
	}

	// effective already excludes everything held out, so put the roster
	// back together before measuring against it. Measuring against
	// livePower alone would let the denominator shrink with every
	// freeze, eroding the floor one validator at a time.
	rosterPower := livePower + heldOutPower()

	// Invariant 1: power floor.
	postFreezePower := livePower - power
	if postFreezePower*100 < rosterPower*autoFloorPct {
		panic(ufmt.Sprintf(
			"auto-freeze would leave %d of %d voting power live, below the %d%% floor",
			postFreezePower, rosterPower, autoFloorPct,
		))
	}

	// Invariant 2: auto-frozen cap.
	if (autoFrozenPower()+power)*100 > rosterPower*autoCapPct {
		panic(ufmt.Sprintf(
			"auto-freeze would put %d of %d voting power in auto-freeze, above the %d%% cap",
			autoFrozenPower()+power, rosterPower, autoCapPct,
		))
	}

	// Invariant 4: expiry, recorded up front so ExpireAutoFrozen needs
	// no knowledge of why the entry was made.
	now := runtime.ChainHeight()
	expiresAt := now + autoFreezeExpiryBlocks

	autoFrozenSet.Set(key, autoFrozenEntry{
		Op:             op,
		SigningAddress: entry.SigningAddress,
		Power:          power,
		Reason:         reason,
		Evidence:       evidence,
		Monitor:        monitorAddr,
		FrozenAt:       now,
		ExpiresAt:      expiresAt,
	})

	publishValsetWithout(cur, entry.SigningAddress)

	chain.Emit(
		"AutoValidatorFrozen",
		"op", key,
		"signingAddr", entry.SigningAddress.String(),
		"monitor", monitorAddr.String(),
		"power", strconv.FormatUint(power, 10),
		"reason", reason,
		"evidence", evidence,
		"expiresAt", strconv.FormatInt(expiresAt, 10),
	)
}

// ExpireAutoFrozen restores every auto-frozen operator whose entry has
// aged past its ExpiresAt. Callable by anyone — it only ever gives
// power back, and the caller pays for the work.
//
// Returns the number of operators restored.
//
// An operator that cannot be restored cleanly — profile gone from
// valoperCache, or KeepRunning flipped to false while frozen — has its
// entry dropped rather than republished, with an AutoValidatorExpired
// event carrying restored=false. Panicking instead would let one stuck
// operator block the sweep for everyone else.
func ExpireAutoFrozen(cur realm) int {
	now := runtime.ChainHeight()

	var due []autoFrozenEntry
	autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
		e := raw.(autoFrozenEntry)
		if now >= e.ExpiresAt {
			due = append(due, e)
		}
		return false
	})
	if len(due) == 0 {
		return 0
	}

	effective := sysparams.GetValsetEffective()
	set := make(map[address]validators.Validator, len(effective)+len(due))
	for _, v := range effective {
		set[v.Address] = v
	}

	restored := 0
	for _, e := range due {
		autoFrozenSet.Remove(e.Op.String())

		// Re-resolve: an operator who rotated while auto-frozen must
		// come back under the CURRENT key, never the captured one.
		raw := valoperCache.Get(e.Op.String())
		if raw == nil {
			emitExpired(e, now, "unknown operator")
			continue
		}
		entry := raw.(cacheEntry)
		if !entry.KeepRunning {
			emitExpired(e, now, "operator opted out")
			continue
		}

		set[entry.SigningAddress] = validators.Validator{
			Address:     entry.SigningAddress,
			PubKey:      entry.SigningPubKey,
			VotingPower: e.Power,
		}
		restored++
		emitExpired(e, now, "")
	}

	if restored > 0 {
		publishValset(cur, set)
	}
	return restored
}

// emitExpired reports one swept entry. skipped is empty when the
// operator was actually restored.
func emitExpired(e autoFrozenEntry, now int64, skipped string) {
	restored := "true"
	if skipped != "" {
		restored = "false"
	}
	chain.Emit(
		"AutoValidatorExpired",
		"op", e.Op.String(),
		"frozenAt", strconv.FormatInt(e.FrozenAt, 10),
		"expiredAt", strconv.FormatInt(now, 10),
		"restored", restored,
		"skipReason", skipped,
	)
}

// ---- monitor management (GovDAO path) ------------------------------------------

// NewRegisterMonitorProposalRequest builds a GovDAO proposal that lets
// addr auto-freeze validators.
func NewRegisterMonitorProposalRequest(cur realm, addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic("invalid address: " + addr.String())
	}

	callback := func(cur realm) error {
		monitors.Set(addr.String(), struct{}{})
		chain.Emit("MonitorRegistered", "addr", addr.String())
		return nil
	}
	return dao.NewProposalRequest(
		"Register validator monitor "+addr.String(),
		"Proposal to let "+addr.String()+" freeze validators automatically, "+
			"subject to the on-chain power floor, auto-frozen cap, protected set and expiry.\n",
		dao.NewSimpleExecutor(0, cur, callback, ""),
	)
}

// NewDeregisterMonitorProposalRequest builds a GovDAO proposal that
// revokes addr's monitor authority. Entries the monitor already made
// are left alone; they expire on their own, or an admin can unfreeze
// them sooner.
func NewDeregisterMonitorProposalRequest(cur realm, addr address) dao.ProposalRequest {
	if !monitors.Has(addr.String()) {
		panic("address is not a registered monitor: " + addr.String())
	}

	callback := func(cur realm) error {
		if _, removed := monitors.Remove(addr.String()); !removed {
			panic("address is not a registered monitor: " + addr.String())
		}
		chain.Emit("MonitorDeregistered", "addr", addr.String())
		return nil
	}
	return dao.NewProposalRequest(
		"Deregister validator monitor "+addr.String(),
		"Proposal to revoke "+addr.String()+"'s automatic freeze authority.\n",
		dao.NewSimpleExecutor(0, cur, callback, ""),
	)
}

// ---- protected set management (GovDAO path) ------------------------------------

// NewAddProtectedProposalRequest builds a GovDAO proposal that makes op
// immune to monitor freezes. Manual freezes are unaffected.
func NewAddProtectedProposalRequest(cur realm, op address) dao.ProposalRequest {
	// Resolve now so a typo'd operator is caught before it consumes a vote.
	mustValoper(op)

	callback := func(cur realm) error {
		protectedSet.Set(op.String(), struct{}{})
		chain.Emit("ProtectedAdded", "op", op.String())
		return nil
	}
	return dao.NewProposalRequest(
		"Protect validator "+op.String()+" from automatic freezes",
		"Proposal to exempt operator "+op.String()+" from monitor-driven freezes. "+
			"GovDAO and freeze admins can still freeze it manually.\n",
		dao.NewSimpleExecutor(0, cur, callback, ""),
	)
}

// NewRemoveProtectedProposalRequest builds a GovDAO proposal that
// removes op from the protected set.
func NewRemoveProtectedProposalRequest(cur realm, op address) dao.ProposalRequest {
	if !protectedSet.Has(op.String()) {
		panic("operator is not in the protected set: " + op.String())
	}

	callback := func(cur realm) error {
		if _, removed := protectedSet.Remove(op.String()); !removed {
			panic("operator is not in the protected set: " + op.String())
		}
		chain.Emit("ProtectedRemoved", "op", op.String())
		return nil
	}
	return dao.NewProposalRequest(
		"Unprotect validator "+op.String(),
		"Proposal to let monitors freeze operator "+op.String()+" again.\n",
		dao.NewSimpleExecutor(0, cur, callback, ""),
	)
}

// ---- read helpers ----------------------------------------------------------------

// IsMonitor returns true if addr may auto-freeze.
func IsMonitor(addr address) bool {
	return monitors.Has(addr.String())
}

// GetMonitors returns every registered monitor address.
func GetMonitors() []address {
	out := make([]address, 0, monitors.Size())
	monitors.Iterate("", "", func(key string, _ any) bool {
		out = append(out, address(key))
		return false
	})
	return out
}

// IsProtected reports whether op is exempt from monitor freezes.
func IsProtected(op address) bool {
	return protectedSet.Has(op.String())
}

// IsAutoFrozen reports whether op is currently auto-frozen.
func IsAutoFrozen(op address) bool {
	return autoFrozenSet.Has(op.String())
}

// GetAutoFrozenOperators returns the operator addresses currently
// auto-frozen.
func GetAutoFrozenOperators() []address {
	out := make([]address, 0, autoFrozenSet.Size())
	autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
		out = append(out, raw.(autoFrozenEntry).Op)
		return false
	})
	return out
}

// ---- render -----------------------------------------------------------------------

// renderAutoFreeze returns the monitor and auto-frozen sections of
// Render. Called from Render in validators.gno.
func renderAutoFreeze() string {
	if monitors.Size() == 0 && autoFrozenSet.Size() == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n## Monitors\n\n")
	if monitors.Size() == 0 {
		sb.WriteString("_(none registered)_\n")
	} else {
		monitors.Iterate("", "", func(key string, _ any) bool {
			sb.WriteString("- " + key + "\n")
			return false
		})
	}

	if autoFrozenSet.Size() > 0 {
		sb.WriteString("\n## Auto-Frozen Validators\n\n")
		sb.WriteString("| Operator | Monitor | Power | Reason | Frozen at | Expires at |\n")
		sb.WriteString("|----------|---------|-------|--------|-----------|------------|\n")
		autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
			e := raw.(autoFrozenEntry)
			sb.WriteString(ufmt.Sprintf("| %s | %s | %d | %s | %d | %d |\n",
				e.Op.String(), e.Monitor.String(), e.Power, e.Reason, e.FrozenAt, e.ExpiresAt))
			return false
		})
	}

	return sb.String()
}

// ---- invariants ---------------------------------------------------------------------

// autoFrozenPower is the total voting power currently held out by the
// monitor path.
func autoFrozenPower() uint64 {
	total := uint64(0)
	autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
		total += raw.(autoFrozenEntry).Power
		return false
	})
	return total
}

// ---- auth ------------------------------------------------------------------------------

// assertMonitorCaller panics unless the call is a direct transaction
// from a registered monitor, and returns that monitor's address.
//
// Same reasoning as assertFreezeAdmin: tx-origin identity would let any
// realm a monitor happens to call freeze validators on its behalf, and
// a monitor is a hot key that signs unattended.
func assertMonitorCaller(_ int, rlm realm) address {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}

	prev := rlm.Previous()
	if !prev.IsUser() {
		panic("monitors must call directly, not through realm " + prev.PkgPath())
	}
	addr := prev.Address()
	if !monitors.Has(addr.String()) {
		panic("caller is not a registered monitor: " + addr.String())
	}
	return addr
}
