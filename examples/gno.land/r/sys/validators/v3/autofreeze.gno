// autofreeze.gno — Phase B: automated validator freeze via off-chain monitors
//
// This is a draft. Phase A (freeze.gno) handles manual freeze/unfreeze by
// GovDAO and a GovDAO-managed admin list. Phase B adds an automated path:
// registered off-chain monitors can propose a freeze with evidence, and the
// chain enforces safety invariants on every auto-freeze attempt.
//
// Design:
//
//   Monitors are registered/deregistered by GovDAO (RegisterMonitor /
//   DeregisterMonitor). Any registered monitor can call AutoFreezeValidator
//   with a reason and evidence string.
//
//   Every AutoFreezeValidator call is guarded by four on-chain invariants:
//
//   1. Power floor: post-freeze live power must be ≥ 80% of total power.
//      (Replaces the count-based 2/3 floor used in Phase A.)
//
//   2. Auto-frozen cap: at most 1/8 of total power may be auto-frozen at once.
//
//   3. Protected set: validators in protectedSet cannot be auto-frozen.
//      (GovDAO manages the set.)
//
//   4. Auto-expiry: auto-frozen entries expire after autoFreezeExpiryBlocks.
//      ExpireAutoFrozen() can be called by anyone to sweep expired entries
//      and restore them automatically.
//
//   The monitor is UNTRUSTED for safety. A compromised monitor's worst-case
//   outcome is bounded and time-limited by invariants 1-4.
//
// Open questions (same as Phase A ADR):
//   - Phase C: require N-of-M monitor attestation before auto-freeze fires.
//   - Rate limiting per monitor (currently no per-monitor rate limit).
//   - Whether auto-unfreeze on expiry is always correct, or should require
//     a human confirmation step.
package validators

import (
	"strconv"
	"strings"

	"chain"
	"chain/runtime"
	runtimeunsafe "chain/runtime/unsafe"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	sysparams "gno.land/r/sys/params"
)

// autoFreezeExpiryBlocks is how many blocks an auto-frozen entry lives before
// it is eligible for auto-unfreeze via ExpireAutoFrozen. ~1 day at 1s/block.
const autoFreezeExpiryBlocks = int64(86400)

// powerFloorPct is the minimum fraction of total power that must remain live
// after an auto-freeze (80%). Stored as a numerator over 100.
const powerFloorPct = 80

// autoFrozenCapPct is the maximum fraction of total power that may be
// auto-frozen at once (12.5% ≈ 1/8). Stored as numerator over 100.
const autoFrozenCapPct = 12

// monitors holds registered monitor addresses (addr.String() → struct{}).
// Only GovDAO can modify this set.
var monitors = avl.NewTree()

// protectedSet holds validator addresses that monitors cannot auto-freeze
// (addr.String() → struct{}). GovDAO manages this set.
var protectedSet = avl.NewTree()

// autoFrozenSet tracks auto-frozen entries separately from manually frozen ones
// so expiry logic only touches auto-freezes.
// key: addr.String() → autoFrozenEntry
var autoFrozenSet = avl.NewTree()

type autoFrozenEntry struct {
	Addr      address
	PubKey    string
	Power     uint64
	Reason    string
	Evidence  string
	Monitor   address
	FrozenAt  int64
	ExpiresAt int64
}

// ---- auto-freeze ------------------------------------------------------------

// AutoFreezeValidator is called by a registered monitor to freeze a validator
// based on observed misbehaviour. All four safety invariants are checked
// before the freeze is applied.
//
// On success the validator is removed from the effective valset and an
// AutoValidatorFrozen event is emitted. The entry expires automatically
// after autoFreezeExpiryBlocks blocks.
func AutoFreezeValidator(cur realm, addr address, reason, evidence string) {
	assertMonitorCaller(cur)

	key := addr.String()

	if IsFrozen(addr) {
		panic("validator is already manually frozen: " + key)
	}
	if _, exists := autoFrozenSet.Get(key); exists {
		panic("validator is already auto-frozen: " + key)
	}
	if _, protected := protectedSet.Get(key); protected {
		panic("validator is in the protected set and cannot be auto-frozen: " + key)
	}

	effective := sysparams.GetValsetEffective()
	var target *validators.Validator
	for i := range effective {
		if effective[i].Address == addr {
			target = &effective[i]
			break
		}
	}
	if target == nil {
		panic("validator not found in effective valset: " + key)
	}

	// Compute total and live power (excluding already-frozen validators).
	totalPower := uint64(0)
	livePower := uint64(0)
	for _, v := range effective {
		totalPower += v.VotingPower
		if !IsFrozen(v.Address) {
			if _, af := autoFrozenSet.Get(v.Address.String()); !af {
				livePower += v.VotingPower
			}
		}
	}

	// Invariant 1: power floor.
	postFreezeLive := livePower - target.VotingPower
	if postFreezeLive*100 < totalPower*powerFloorPct {
		panic(ufmt.Sprintf(
			"auto-freeze would drop live power to %d/%d (%d%%), below the %d%% floor",
			postFreezeLive, totalPower, postFreezeLive*100/totalPower, powerFloorPct,
		))
	}

	// Invariant 2: auto-frozen cap.
	autoFrozenPower := uint64(0)
	autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
		e := raw.(autoFrozenEntry)
		autoFrozenPower += e.Power
		return false
	})
	if (autoFrozenPower+target.VotingPower)*100 > totalPower*autoFrozenCapPct {
		panic(ufmt.Sprintf(
			"auto-freeze would push auto-frozen power to %d/%d, exceeding the %d%% cap",
			autoFrozenPower+target.VotingPower, totalPower, autoFrozenCapPct,
		))
	}

	monitorAddr := runtimeunsafe.OriginCaller()
	now := runtime.ChainHeight()

	autoFrozenSet.Set(key, autoFrozenEntry{
		Addr:      addr,
		PubKey:    target.PubKey,
		Power:     target.VotingPower,
		Reason:    reason,
		Evidence:  evidence,
		Monitor:   monitorAddr,
		FrozenAt:  now,
		ExpiresAt: now + autoFreezeExpiryBlocks,
	})

	publishValsetWithout(cur, addr)

	chain.Emit(
		"AutoValidatorFrozen",
		"addr", key,
		"monitor", monitorAddr.String(),
		"reason", reason,
		"expiresAt", strconv.FormatInt(now+autoFreezeExpiryBlocks, 10),
	)
}

// ExpireAutoFrozen sweeps all auto-frozen entries whose ExpiresAt ≤ current
// height and restores them to the effective valset. Callable by anyone.
//
// Returns the number of entries restored.
func ExpireAutoFrozen(cur realm) int {
	now := runtime.ChainHeight()
	var toExpire []autoFrozenEntry

	autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
		e := raw.(autoFrozenEntry)
		if now >= e.ExpiresAt {
			toExpire = append(toExpire, e)
		}
		return false
	})

	if len(toExpire) == 0 {
		return 0
	}

	effective := sysparams.GetValsetEffective()
	set := make(map[address]validators.Validator, len(effective)+len(toExpire))
	for _, v := range effective {
		set[v.Address] = v
	}

	for _, e := range toExpire {
		autoFrozenSet.Remove(e.Addr.String())
		set[e.Addr] = validators.Validator{
			Address:     e.Addr,
			PubKey:      e.PubKey,
			VotingPower: e.Power,
		}
		chain.Emit(
			"AutoValidatorExpired",
			"addr", e.Addr.String(),
			"frozenAt", strconv.FormatInt(e.FrozenAt, 10),
			"expiredAt", strconv.FormatInt(now, 10),
		)
	}

	publishValsetFromMap(cur, set)
	return len(toExpire)
}

// ---- monitor management (GovDAO only) ---------------------------------------

// RegisterMonitor adds addr to the set of trusted monitors. GovDAO only.
func RegisterMonitor(cur realm, addr address) {
	assertGovDAOCaller(cur)
	monitors.Set(addr.String(), struct{}{})
	chain.Emit("MonitorRegistered", "addr", addr.String())
}

// DeregisterMonitor removes addr from the monitor set. GovDAO only.
func DeregisterMonitor(cur realm, addr address) {
	assertGovDAOCaller(cur)
	if _, removed := monitors.Remove(addr.String()); !removed {
		panic("address not in monitor set: " + addr.String())
	}
	chain.Emit("MonitorDeregistered", "addr", addr.String())
}

// IsMonitor returns true if addr is a registered monitor.
func IsMonitor(addr address) bool {
	_, ok := monitors.Get(addr.String())
	return ok
}

// GetMonitors returns all registered monitor addresses.
func GetMonitors() []address {
	out := make([]address, 0, monitors.Size())
	monitors.Iterate("", "", func(key string, _ any) bool {
		out = append(out, address(key))
		return false
	})
	return out
}

// ---- protected set management (GovDAO only) ---------------------------------

// AddProtected adds addr to the set of validators that monitors cannot
// auto-freeze. GovDAO only.
func AddProtected(cur realm, addr address) {
	assertGovDAOCaller(cur)
	protectedSet.Set(addr.String(), struct{}{})
	chain.Emit("ProtectedAdded", "addr", addr.String())
}

// RemoveProtected removes addr from the protected set. GovDAO only.
func RemoveProtected(cur realm, addr address) {
	assertGovDAOCaller(cur)
	if _, removed := protectedSet.Remove(addr.String()); !removed {
		panic("address not in protected set: " + addr.String())
	}
	chain.Emit("ProtectedRemoved", "addr", addr.String())
}

// IsProtected returns true if addr is in the protected set.
func IsProtected(addr address) bool {
	_, ok := protectedSet.Get(addr.String())
	return ok
}

// ---- read helpers -----------------------------------------------------------

// IsAutoFrozen returns true if addr is currently auto-frozen.
func IsAutoFrozen(addr address) bool {
	_, ok := autoFrozenSet.Get(addr.String())
	return ok
}

// GetAutoFrozenValidators returns the addresses of all auto-frozen validators.
func GetAutoFrozenValidators() []address {
	out := make([]address, 0, autoFrozenSet.Size())
	autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
		e := raw.(autoFrozenEntry)
		out = append(out, e.Addr)
		return false
	})
	return out
}

// ---- render extension -------------------------------------------------------

// renderAutoFreeze appends auto-freeze status to Render output.
func renderAutoFreeze() string {
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
		sb.WriteString("| Address | Monitor | Reason | Frozen At | Expires At |\n")
		sb.WriteString("|---------|---------|--------|-----------|------------|\n")
		autoFrozenSet.Iterate("", "", func(_ string, raw any) bool {
			e := raw.(autoFrozenEntry)
			sb.WriteString(ufmt.Sprintf("| %s | %s | %s | %d | %d |\n",
				e.Addr.String(), e.Monitor.String(), e.Reason, e.FrozenAt, e.ExpiresAt))
			return false
		})
	}

	return sb.String()
}

// ---- internal auth ----------------------------------------------------------

// assertMonitorCaller panics if the OriginCaller is not a registered monitor.
func assertMonitorCaller(_ realm) {
	caller := runtimeunsafe.OriginCaller()
	if _, ok := monitors.Get(caller.String()); !ok {
		panic("caller is not a registered monitor: " + caller.String())
	}
}
