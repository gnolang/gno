// freeze.gno — Phase A: manual validator freeze/unfreeze
//
// Authority model:
//
//	Both FreezeValidator and UnfreezeValidator are callable by two paths:
//
//	1. GovDAO proposal (always). Any GovDAO member can propose a freeze or
//	   unfreeze via NewFreezeProposalRequest / NewUnfreezeProposalRequest.
//
//	2. Direct call by an address in the freezeAdmins list. The list is
//	   managed exclusively by GovDAO (AddFreezeAdmin / RemoveFreezeAdmin).
//	   The list may be empty, in which case only path 1 is available.
//
// The liveness floor is count-based for this draft (≥2/3 of total
// validators must remain live after a freeze). Phase B will switch to a
// power-weighted floor.
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

const govDAORealmPath = "gno.land/r/gov/dao"

// freezeAdmins holds addresses authorised to call FreezeValidator or
// UnfreezeValidator directly (without a GovDAO proposal). The set is
// managed by GovDAO and may be empty.
//
// key: addr.String() → struct{}
var freezeAdmins = avl.NewTree()

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

// ---- freeze/unfreeze --------------------------------------------------------

// FreezeValidator removes addr from the effective valset and records its
// power for later restore. Callable by any address in freezeAdmins, or
// via a GovDAO executor (use NewFreezeProposalRequest for the latter).
//
// Panics if addr is not in the effective valset, if it is already frozen,
// or if the freeze would violate the liveness floor.
func FreezeValidator(cur realm, addr address, reason string) {
	assertFreezeOrAdmin(cur)
	doFreeze(cur, addr, reason)
}

// NewFreezeProposalRequest builds a GovDAO proposal that, when executed,
// freezes addr. Use this when no admin is configured or when governance
// sign-off is desired regardless.
func NewFreezeProposalRequest(cur realm, addr address, reason string) dao.ProposalRequest {
	title := "Freeze validator " + addr.String()
	desc := "Proposal to remove validator " + addr.String() + " from the active set.\n\n**Reason:** " + reason + "\n"

	callback := func(cur realm) error {
		doFreeze(cur, addr, reason)
		return nil
	}
	return dao.NewProposalRequest(title, desc, dao.NewSimpleExecutor(0, cur, callback, ""))
}

// UnfreezeValidator restores addr to the effective valset with its saved
// power. Callable by any address in freezeAdmins, or via a GovDAO executor
// (use NewUnfreezeProposalRequest for the latter).
func UnfreezeValidator(cur realm, addr address, reason string) {
	assertFreezeOrAdmin(cur)
	doUnfreeze(cur, addr, reason)
}

// NewUnfreezeProposalRequest builds a GovDAO proposal that, when executed,
// restores addr to the effective valset with its saved power.
func NewUnfreezeProposalRequest(cur realm, addr address, reason string) dao.ProposalRequest {
	key := addr.String()
	raw := frozenSet.Get(key)
	if raw == nil {
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
		doUnfreeze(cur, addr, reason)
		return nil
	}
	return dao.NewProposalRequest(title, desc.String(), dao.NewSimpleExecutor(0, cur, callback, ""))
}

// ---- admin list management (GovDAO only) ------------------------------------

// AddFreezeAdmin adds addr to the freeze admin list. GovDAO only.
func AddFreezeAdmin(cur realm, addr address) {
	assertGovDAOCaller(cur)
	freezeAdmins.Set(addr.String(), struct{}{})
	chain.Emit("FreezeAdminAdded", "addr", addr.String())
}

// RemoveFreezeAdmin removes addr from the freeze admin list. GovDAO only.
func RemoveFreezeAdmin(cur realm, addr address) {
	assertGovDAOCaller(cur)
	if _, removed := freezeAdmins.Remove(addr.String()); !removed {
		panic("address not in freeze admin list: " + addr.String())
	}
	chain.Emit("FreezeAdminRemoved", "addr", addr.String())
}

// IsFreezeAdmin returns true if addr is in the freeze admin list.
func IsFreezeAdmin(addr address) bool {
	ok := freezeAdmins.Has(addr.String())
	return ok
}

// GetFreezeAdmins returns all addresses in the freeze admin list.
func GetFreezeAdmins() []address {
	out := make([]address, 0, freezeAdmins.Size())
	freezeAdmins.Iterate("", "", func(key string, _ any) bool {
		out = append(out, address(key))
		return false
	})
	return out
}

// ---- read helpers -----------------------------------------------------------

// IsFrozen returns true if addr is currently frozen.
func IsFrozen(addr address) bool {
	exists := frozenSet.Has(addr.String())
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

// ---- render -----------------------------------------------------------------

// renderFrozen returns a markdown section listing frozen validators and
// current freeze admins. Called by Render in validators.gno.
func renderFrozen() string {
	var sb strings.Builder

	if frozenSet.Size() > 0 {
		sb.WriteString("\n## Frozen Validators\n\n")
		sb.WriteString("| Address | Reason | Frozen At |\n")
		sb.WriteString("|---------|--------|----------|\n")
		frozenSet.Iterate("", "", func(key string, raw any) bool {
			e := raw.(frozenEntry)
			sb.WriteString(ufmt.Sprintf("| %s | %s | %d |\n", key, e.Reason, e.FrozenAt))
			return false
		})
	}

	sb.WriteString("\n## Freeze Admins\n\n")
	if freezeAdmins.Size() == 0 {
		sb.WriteString("_(none — only GovDAO can freeze/unfreeze)_\n")
	} else {
		freezeAdmins.Iterate("", "", func(key string, _ any) bool {
			sb.WriteString("- " + key + "\n")
			return false
		})
	}

	return sb.String()
}

// ---- internal helpers -------------------------------------------------------

// doFreeze is the shared freeze implementation (auth already checked by caller).
func doFreeze(cur realm, addr address, reason string) {
	key := addr.String()
	if frozenSet.Has(key) {
		panic("validator is already frozen: " + key)
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

	// Liveness floor: count-based for Phase A.
	// Phase B will use a power-weighted floor (e.g. 80% of total power).
	live := len(effective) - frozenSet.Size() - 1
	if live*3 < len(effective)*2 {
		panic(ufmt.Sprintf(
			"freeze would drop live validators to %d/%d, violating the 2/3 floor",
			live, len(effective),
		))
	}

	frozenSet.Set(key, frozenEntry{
		Addr:        addr,
		PubKey:      target.PubKey,
		VotingPower: target.VotingPower,
		Reason:      reason,
		FrozenAt:    runtime.ChainHeight(),
	})

	publishValsetWithout(cur, addr)

	chain.Emit(
		"ValidatorFrozen",
		"addr", key,
		"reason", reason,
		"height", strconv.FormatInt(runtime.ChainHeight(), 10),
	)
}

// doUnfreeze is the shared unfreeze implementation (auth already checked).
func doUnfreeze(cur realm, addr address, reason string) {
	key := addr.String()
	raw := frozenSet.Get(key)
	if raw == nil {
		panic("validator is not frozen: " + key)
	}
	e := raw.(frozenEntry)
	frozenSet.Remove(key)

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
}

// assertFreezeOrAdmin allows the call through if either:
//   - the immediate calling realm is GovDAO (proposal executor path), or
//   - the transaction origin caller is in freezeAdmins (direct admin path).
func assertFreezeOrAdmin(cur realm) {
	if cur.Previous().PkgPath() == govDAORealmPath {
		return
	}
	caller := runtimeunsafe.OriginCaller()
	if freezeAdmins.Has(caller.String()) {
		return
	}
	panic("caller is not a freeze admin and this is not a GovDAO call; " +
		"use NewFreezeProposalRequest or be added to the admin list via AddFreezeAdmin")
}

// assertGovDAOCaller panics if the immediate calling realm is not GovDAO.
func assertGovDAOCaller(cur realm) {
	if cur.Previous().PkgPath() != govDAORealmPath {
		panic("must be called from " + govDAORealmPath)
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
