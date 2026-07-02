// freeze.gno — Phase A: manual validator freeze/unfreeze
//
// This is an early draft exploring what a freeze mechanism could look
// like for gno.land's PoA validator set. The intent is to open a
// discussion on authority, lifecycle, and liveness constraints before
// settling on a final design.
//
// Design sketch (Phase A — manual only):
//
//   FreezeValidator(addr, reason) — callable by freezeAdmin.
//     Records the validator's current power, removes it from the
//     effective valset, and emits a ValidatorFrozen event. The node
//     drops out of consensus on the next block without any new
//     consensus-level code.
//
//   NewUnfreezeProposalRequest(...) — builds a GovDAO proposal that,
//     when executed, restores the validator with its saved power and
//     emits ValidatorUnfrozen.
//
//   SetFreezeAdmin(addr) — GovDAO-only; updates who can freeze.
//
// Open questions worth discussing before stabilising:
//   - Should FreezeValidator be GovDAO-gated instead of admin-gated?
//   - Auto-expiry: frozen entries could self-expire after N blocks.
//   - Phase B: off-chain monitor + on-chain liveness floor; see ADR.
//   - Slash: no bonded stake today; punitive mechanism deferred.
package validators

import (
	"sort"
	"strconv"
	"strings"

	"chain"
	"chain/runtime"
	runtimeunsafe "chain/runtime/unsafe"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

// freezeAdmin is the address authorised to call FreezeValidator
// without a GovDAO proposal. It is expected to be a multisig or
// similarly guarded address. GovDAO can rotate it via SetFreezeAdmin.
//
// Set to the zero address by default so the feature is inert until
// governance explicitly configures an admin.
var freezeAdmin address

// frozenEntry records the state of a frozen validator so it can be
// restored on unfreeze.
type frozenEntry struct {
	Addr        address
	PubKey      string
	VotingPower uint64
	Reason      string
	FrozenAt    int64
}

// frozenSet maps signing address (string) → frozenEntry.
var frozenSet = avl.NewTree()

// FreezeValidator removes addr from the effective valset and records
// its power so it can be restored later. Callable only by freezeAdmin.
//
// Panics if addr is not currently in the effective valset, if addr is
// already frozen, or if the freeze would violate the liveness floor
// (fewer than 2/3 of validators remaining — very conservative for now).
func FreezeValidator(cur realm, addr address, reason string) {
	assertFreezeAdmin(cur)

	key := addr.String()
	if _, exists := frozenSet.Get(key); exists {
		panic("validator is already frozen: " + key)
	}

	// Find addr in the effective valset.
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

	// Liveness floor: refuse if freezing would leave fewer than ⌈2/3⌉
	// of total validators active. Kept intentionally conservative for
	// this draft; Phase B will use a power-weighted floor instead.
	live := len(effective) - frozenSet.Size() - 1 // -1 for the one we're about to freeze
	if live*3 < len(effective)*2 {
		panic(ufmt.Sprintf(
			"freeze would drop live validators to %d/%d, violating the 2/3 floor",
			live, len(effective),
		))
	}

	// Record the frozen entry.
	frozenSet.Set(key, frozenEntry{
		Addr:        addr,
		PubKey:      target.PubKey,
		VotingPower: target.VotingPower,
		Reason:      reason,
		FrozenAt:    runtime.ChainHeight(),
	})

	// Publish the updated valset (addr removed, power=0).
	publishValsetWithout(cur, addr)

	chain.Emit(
		"ValidatorFrozen",
		"addr", key,
		"reason", reason,
		"height", strconv.FormatInt(runtime.ChainHeight(), 10),
	)
}

// NewUnfreezeProposalRequest builds a GovDAO proposal that, when
// executed, restores addr to the effective valset with its saved power.
func NewUnfreezeProposalRequest(cur realm, addr address, reason string) dao.ProposalRequest {
	key := addr.String()
	raw, exists := frozenSet.Get(key)
	if !exists {
		panic("validator is not frozen: " + key)
	}
	entry := raw.(frozenEntry)

	title := "Unfreeze validator " + key
	var desc strings.Builder
	desc.WriteString("Proposal to restore validator ")
	desc.WriteString(key)
	desc.WriteString(" to the active set.\n\n")
	desc.WriteString("- **Frozen at height:** ")
	desc.WriteString(strconv.FormatInt(entry.FrozenAt, 10))
	desc.WriteString("\n- **Freeze reason:** ")
	desc.WriteString(entry.Reason)
	desc.WriteString("\n- **Unfreeze reason:** ")
	desc.WriteString(reason)
	desc.WriteString("\n")

	callback := func(cur realm) error {
		raw2, exists2 := frozenSet.Get(key)
		if !exists2 {
			panic("validator is not frozen at execution time: " + key)
		}
		e := raw2.(frozenEntry)
		frozenSet.Remove(key)

		// Re-add to the effective valset.
		effective := sysparams.GetValsetEffective()
		set := make(map[address]validators.Validator, len(effective)+1)
		for _, v := range effective {
			set[v.Address] = v
		}
		set[addr] = validators.Validator{
			Address:     addr,
			PubKey:      e.PubKey,
			VotingPower: e.VotingPower,
		}

		publishValsetFromMap(cur, set)

		chain.Emit(
			"ValidatorUnfrozen",
			"addr", key,
			"reason", reason,
			"height", strconv.FormatInt(runtime.ChainHeight(), 10),
		)
		return nil
	}

	return dao.NewProposalRequest(title, desc.String(), dao.NewSimpleExecutor(0, cur, callback, ""))
}

// SetFreezeAdmin updates the freeze admin address. Must be called via
// a GovDAO proposal (enforced by assertGovDAOCaller).
func SetFreezeAdmin(cur realm, addr address) {
	assertGovDAOCaller(cur)
	freezeAdmin = addr
	chain.Emit("FreezeAdminUpdated", "addr", addr.String())
}

// GetFreezeAdmin returns the current freeze admin address.
func GetFreezeAdmin() address {
	return freezeAdmin
}

// IsFrozen returns true if addr is currently frozen.
func IsFrozen(addr address) bool {
	_, exists := frozenSet.Get(addr.String())
	return exists
}

// GetFrozenValidators returns the list of currently frozen validator addresses.
func GetFrozenValidators() []address {
	out := make([]address, 0, frozenSet.Size())
	frozenSet.Iterate("", "", func(_ string, raw any) bool {
		e := raw.(frozenEntry)
		out = append(out, e.Addr)
		return false
	})
	return out
}

// renderFrozen returns a markdown section listing frozen validators,
// or an empty string if none are frozen. Called by Render.
func renderFrozen() string {
	if frozenSet.Size() == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n## Frozen Validators\n\n")
	sb.WriteString("| Address | Reason | Frozen At |\n")
	sb.WriteString("|---------|--------|----------|\n")
	frozenSet.Iterate("", "", func(key string, raw any) bool {
		e := raw.(frozenEntry)
		sb.WriteString(ufmt.Sprintf("| %s | %s | %d |\n", key, e.Reason, e.FrozenAt))
		return false
	})
	return sb.String()
}

// assertFreezeAdmin panics if the transaction origin caller is not freezeAdmin.
func assertFreezeAdmin(_ realm) {
	if freezeAdmin == address("") {
		panic("freeze admin not configured; set via SetFreezeAdmin GovDAO proposal")
	}
	caller := runtimeunsafe.OriginCaller()
	if caller != freezeAdmin {
		panic("caller " + caller.String() + " is not the freeze admin")
	}
}

// assertGovDAOCaller panics if the immediate calling realm is not GovDAO.
// Only used for SetFreezeAdmin.
const govDAORealmPath = "gno.land/r/gov/dao"

func assertGovDAOCaller(cur realm) {
	if cur.Previous().PkgPath() != govDAORealmPath {
		panic("SetFreezeAdmin must be called from " + govDAORealmPath)
	}
}

// publishValsetWithout rebuilds and publishes the effective valset
// excluding addr (freeze path).
func publishValsetWithout(cur realm, addr address) {
	effective := sysparams.GetValsetEffective()
	set := make(map[address]validators.Validator, len(effective))
	for _, v := range effective {
		if v.Address != addr {
			set[v.Address] = v
		}
	}
	publishValsetFromMap(cur, set)
}

// publishValsetFromMap encodes a map[address]Validator into the
// sysparams wire format and calls SetValsetProposal.
func publishValsetFromMap(cur realm, set map[address]validators.Validator) {
	entries := make([]string, 0, len(set))
	for _, v := range set {
		entries = append(entries, v.PubKey+":"+strconv.FormatUint(v.VotingPower, 10))
	}
	sort.Strings(entries)
	sysparams.SetValsetProposal(cross(cur), entries)
}
