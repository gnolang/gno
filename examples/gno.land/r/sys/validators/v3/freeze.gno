// freeze.gno — Phase A: manual validator freeze/unfreeze.
//
// A freeze holds an operator out of the effective valset without
// forgetting them: the pre-freeze voting power is recorded so an
// unfreeze can put them back. It is the emergency lever for "this
// validator is misbehaving right now and governance is too slow".
//
// Authority model — two paths, both always available:
//
//  1. GovDAO proposal. NewFreezeProposalRequest /
//     NewUnfreezeProposalRequest build the requests; GovDAO executes.
//
//  2. Direct call by an address in freezeAdmins. The list is managed
//     exclusively by GovDAO and starts empty, so the realm is in
//     GovDAO-only mode until governance opts into a faster path.
//
// Operator-keyed, like ValoperChange. Callers name an OPERATOR
// address; the signing key is resolved from valoperCache at freeze
// time and RE-resolved at unfreeze time. Keying on the signing
// address instead would restore an operator who rotated while frozen
// under their retired pubkey — exactly the stale-key hazard
// newValoperChangeExecutor documents and defends against.
//
// FREE-FORM STRINGS: every caller-supplied reason (and, in Phase B,
// evidence) is escaped at each markdown sink it reaches — TableCell for
// Render's table cells, InlineText for GovDAO proposal descriptions,
// which GovDAO renders raw. Escaping happens at the sink, never on the
// way into state: stored values and chain.Emit payloads must stay
// verbatim, and sanitize is not idempotent, so a value that could be
// escaped twice is wrong by construction. Any new sink for one of these
// fields needs its own escape.
package validators

import (
	"sort"
	"strconv"
	"strings"

	"chain"
	"chain/runtime"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

// Liveness floor: at least freezeFloorNum/freezeFloorDen of the full
// validator roster — by count AND by voting power — must remain live
// after a freeze. Phase B tightens this further for the untrusted
// monitor-driven path.
const (
	freezeFloorNum = 2
	freezeFloorDen = 3
)

// freezeAdmins holds addresses authorised to call FreezeValidator or
// UnfreezeValidator directly, without a GovDAO proposal. Managed
// exclusively by GovDAO; empty by default.
//
// key: addr.String() → struct{}
var freezeAdmins = avl.NewTree()

// frozenSet maps OPERATOR address → frozenEntry.
var frozenSet = avl.NewTree()

// frozenEntry records what a freeze took away, so an unfreeze can give
// it back. Only VotingPower is authoritative on restore: the signing
// key is re-resolved from valoperCache at unfreeze time, and
// SigningAddress is kept purely as an audit trail of what was pulled.
type frozenEntry struct {
	Op             address
	SigningAddress address
	VotingPower    uint64
	Reason         string
	FrozenAt       int64
}

// ---- freeze/unfreeze (admin path) -------------------------------------------

// FreezeValidator removes op from the effective valset and records its
// power for a later restore. Callable only by an address in
// freezeAdmins, as a direct user transaction.
//
// Panics if op is unknown, not currently in the effective valset,
// already frozen, or if the freeze would breach the liveness floor.
func FreezeValidator(cur realm, op address, reason string) {
	assertFreezeAdmin(0, cur)
	doFreeze(cur, op, reason)
}

// UnfreezeValidator restores op to the effective valset at its recorded
// power, under its CURRENT signing key. Callable only by an address in
// freezeAdmins, as a direct user transaction.
func UnfreezeValidator(cur realm, op address, reason string) {
	assertFreezeAdmin(0, cur)
	doUnfreeze(cur, op, reason)
}

// ---- freeze/unfreeze (GovDAO path) ------------------------------------------

// NewFreezeProposalRequest builds a GovDAO proposal that freezes op on
// execution. Membership and liveness are re-checked at execution time,
// so a proposal that was valid when filed still fails safely if the
// valset moved underneath it.
func NewFreezeProposalRequest(cur realm, op address, reason string) dao.ProposalRequest {
	// Creation-time validation, mirroring NewValidatorProposalRequest:
	// reject obviously-doomed proposals before they consume a vote.
	entry := mustValoper(op)
	if frozenSet.Has(op.String()) {
		panic("operator is already frozen: " + op.String())
	}

	title := "Freeze validator " + op.String()
	var desc strings.Builder
	desc.WriteString("Proposal to hold operator ")
	desc.WriteString(op.String())
	desc.WriteString(" out of the active validator set.\n\n")
	desc.WriteString("- **Signing address:** ")
	desc.WriteString(entry.SigningAddress.String())
	desc.WriteString("\n- **Reason:** ")
	desc.WriteString(sanitize.InlineText(reason))
	desc.WriteString("\n")

	callback := func(cur realm) error {
		doFreeze(cur, op, reason)
		return nil
	}
	return dao.NewProposalRequest(title, desc.String(), dao.NewSimpleExecutor(0, cur, callback, ""))
}

// NewUnfreezeProposalRequest builds a GovDAO proposal that restores op
// to the effective valset at its recorded power.
func NewUnfreezeProposalRequest(cur realm, op address, reason string) dao.ProposalRequest {
	e := mustFrozen(op)

	title := "Unfreeze validator " + op.String()
	var desc strings.Builder
	desc.WriteString("Proposal to restore operator ")
	desc.WriteString(op.String())
	desc.WriteString(" to the active validator set at power ")
	desc.WriteString(strconv.FormatUint(e.VotingPower, 10))
	desc.WriteString(".\n\n")
	desc.WriteString("- **Frozen at height:** ")
	desc.WriteString(strconv.FormatInt(e.FrozenAt, 10))
	desc.WriteString("\n- **Freeze reason:** ")
	desc.WriteString(sanitize.InlineText(e.Reason))
	desc.WriteString("\n- **Unfreeze reason:** ")
	desc.WriteString(sanitize.InlineText(reason))
	desc.WriteString("\n\nThe operator's signing key is re-resolved at execution time, ")
	desc.WriteString("so a key rotated during the freeze is honoured.\n")

	callback := func(cur realm) error {
		doUnfreeze(cur, op, reason)
		return nil
	}
	return dao.NewProposalRequest(title, desc.String(), dao.NewSimpleExecutor(0, cur, callback, ""))
}

// NewDropFrozenProposalRequest builds a GovDAO proposal that forgets a
// frozen entry WITHOUT restoring the operator — the permanent-removal
// exit from a freeze.
//
// It exists because a frozen operator is otherwise stuck: they are
// absent from the effective valset, so a {op, Power: 0} ValoperChange
// panics with "validator does not exist", and unfreeze would be the
// only way out. Dropping the entry leaves the operator out of the
// valset for good; re-admission goes through the normal
// NewValidatorProposalRequest add path.
func NewDropFrozenProposalRequest(cur realm, op address, reason string) dao.ProposalRequest {
	e := mustFrozen(op)

	title := "Drop frozen validator " + op.String()
	var desc strings.Builder
	desc.WriteString("Proposal to permanently drop frozen operator ")
	desc.WriteString(op.String())
	desc.WriteString(" without restoring it to the active set.\n\n")
	desc.WriteString("- **Frozen at height:** ")
	desc.WriteString(strconv.FormatInt(e.FrozenAt, 10))
	desc.WriteString("\n- **Freeze reason:** ")
	desc.WriteString(sanitize.InlineText(e.Reason))
	desc.WriteString("\n- **Drop reason:** ")
	desc.WriteString(sanitize.InlineText(reason))
	desc.WriteString("\n")

	callback := func(cur realm) error {
		doDropFrozen(op, reason)
		return nil
	}
	return dao.NewProposalRequest(title, desc.String(), dao.NewSimpleExecutor(0, cur, callback, ""))
}

// ---- admin list management (GovDAO path) ------------------------------------

// NewAddFreezeAdminProposalRequest builds a GovDAO proposal that adds
// addr to the freeze admin list.
func NewAddFreezeAdminProposalRequest(cur realm, addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic("invalid address: " + addr.String())
	}

	callback := func(cur realm) error {
		freezeAdmins.Set(addr.String(), struct{}{})
		chain.Emit("FreezeAdminAdded", "addr", addr.String())
		return nil
	}
	return dao.NewProposalRequest(
		"Add freeze admin "+addr.String(),
		"Proposal to let "+addr.String()+" freeze and unfreeze validators without a GovDAO vote.\n",
		dao.NewSimpleExecutor(0, cur, callback, ""),
	)
}

// NewRemoveFreezeAdminProposalRequest builds a GovDAO proposal that
// removes addr from the freeze admin list.
func NewRemoveFreezeAdminProposalRequest(cur realm, addr address) dao.ProposalRequest {
	if !freezeAdmins.Has(addr.String()) {
		panic("address not in freeze admin list: " + addr.String())
	}

	callback := func(cur realm) error {
		if _, removed := freezeAdmins.Remove(addr.String()); !removed {
			panic("address not in freeze admin list: " + addr.String())
		}
		chain.Emit("FreezeAdminRemoved", "addr", addr.String())
		return nil
	}
	return dao.NewProposalRequest(
		"Remove freeze admin "+addr.String(),
		"Proposal to revoke "+addr.String()+"'s direct freeze/unfreeze authority.\n",
		dao.NewSimpleExecutor(0, cur, callback, ""),
	)
}

// ---- read helpers -----------------------------------------------------------

// IsFreezeAdmin returns true if addr may freeze/unfreeze directly.
func IsFreezeAdmin(addr address) bool {
	return freezeAdmins.Has(addr.String())
}

// GetFreezeAdmins returns every address in the freeze admin list.
func GetFreezeAdmins() []address {
	out := make([]address, 0, freezeAdmins.Size())
	freezeAdmins.Iterate("", "", func(key string, _ any) bool {
		out = append(out, address(key))
		return false
	})
	return out
}

// IsFrozen reports whether the OPERATOR op is currently frozen.
func IsFrozen(op address) bool {
	return frozenSet.Has(op.String())
}

// GetFrozenOperators returns the operator addresses currently frozen.
func GetFrozenOperators() []address {
	out := make([]address, 0, frozenSet.Size())
	frozenSet.Iterate("", "", func(_ string, raw any) bool {
		out = append(out, raw.(frozenEntry).Op)
		return false
	})
	return out
}

// ---- render -----------------------------------------------------------------

// renderFrozen returns the frozen-validator and freeze-admin sections of
// Render. Called from Render in validators.gno.
func renderFrozen() string {
	// Nothing to say on a chain that has never used the lever — keep the
	// default valset render free of boilerplate.
	if frozenSet.Size() == 0 && freezeAdmins.Size() == 0 {
		return ""
	}

	var sb strings.Builder

	if frozenSet.Size() > 0 {
		sb.WriteString("\n## Frozen Validators\n\n")
		sb.WriteString("| Operator | Signing address | Power | Reason | Frozen at |\n")
		sb.WriteString("|----------|-----------------|-------|--------|----------|\n")
		frozenSet.Iterate("", "", func(_ string, raw any) bool {
			e := raw.(frozenEntry)
			// Escaped even though the writer is privileged: a bare `|`
			// alone silently reshapes the table. See the file header.
			sb.WriteString(ufmt.Sprintf("| %s | %s | %d | %s | %d |\n",
				e.Op.String(), e.SigningAddress.String(), e.VotingPower,
				sanitize.TableCell(e.Reason), e.FrozenAt))
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

// ---- core -------------------------------------------------------------------

// doFreeze applies a freeze. Auth is the caller's responsibility:
// FreezeValidator gates on freezeAdmins, and the proposal callback is
// reachable only through GovDAO execution.
func doFreeze(cur realm, op address, reason string) {
	key := op.String()
	if frozenSet.Has(key) {
		panic("operator is already frozen: " + key)
	}

	entry := mustValoper(op)
	effective := sysparams.GetValsetEffective()

	power, live := uint64(0), false
	for _, v := range effective {
		if v.Address == entry.SigningAddress {
			power, live = v.VotingPower, true
			break
		}
	}
	if !live {
		panic("operator is not in the effective valset: " + key)
	}

	assertLivenessFloor(effective, power)

	now := runtime.ChainHeight()
	frozenSet.Set(key, frozenEntry{
		Op:             op,
		SigningAddress: entry.SigningAddress,
		VotingPower:    power,
		Reason:         reason,
		FrozenAt:       now,
	})

	publishValsetWithout(cur, entry.SigningAddress)

	chain.Emit(
		"ValidatorFrozen",
		"op", key,
		"signingAddr", entry.SigningAddress.String(),
		"power", strconv.FormatUint(power, 10),
		"reason", reason,
		"height", strconv.FormatInt(now, 10),
	)
}

// doUnfreeze restores a frozen operator at its recorded power, under
// whatever signing key valoperCache holds RIGHT NOW.
func doUnfreeze(cur realm, op address, reason string) {
	key := op.String()
	e := mustFrozen(op)

	// Re-resolve rather than trusting e.SigningAddress: an operator that
	// rotated while frozen must come back under the new key. Republishing
	// the captured one would hand consensus a retired pubkey — the key
	// they rotated away from, possibly because it leaked.
	entry := mustValoper(op)

	// KeepRunning=false is a binding opt-out, same as in
	// newValoperChangeExecutor. An operator who opted out while frozen
	// does not get silently reinstated.
	if !entry.KeepRunning {
		panic("operator " + key + " has KeepRunning=false; refusing to restore " +
			"(operator must call UpdateKeepRunning(true), or GovDAO can forget the " +
			"entry via NewDropFrozenProposalRequest)")
	}

	frozenSet.Remove(key)

	effective := sysparams.GetValsetEffective()
	set := make(map[address]validators.Validator, len(effective)+1)
	for _, v := range effective {
		set[v.Address] = v
	}
	set[entry.SigningAddress] = validators.Validator{
		Address:     entry.SigningAddress,
		PubKey:      entry.SigningPubKey,
		VotingPower: e.VotingPower,
	}

	publishValset(cur, set)

	chain.Emit(
		"ValidatorUnfrozen",
		"op", key,
		"signingAddr", entry.SigningAddress.String(),
		"power", strconv.FormatUint(e.VotingPower, 10),
		"reason", reason,
		"height", strconv.FormatInt(runtime.ChainHeight(), 10),
	)
}

// doDropFrozen forgets a frozen entry without touching the valset. The
// operator is already out; this makes it permanent.
func doDropFrozen(op address, reason string) {
	key := op.String()
	if _, removed := frozenSet.Remove(key); !removed {
		panic("operator is not frozen: " + key)
	}

	chain.Emit(
		"FrozenValidatorDropped",
		"op", key,
		"reason", reason,
		"height", strconv.FormatInt(runtime.ChainHeight(), 10),
	)
}

// ---- invariants -------------------------------------------------------------

// heldOutBy names the mechanism currently holding op out of the
// effective valset, or "" if nothing is.
//
// Valset-membership changes must refuse a held-out operator. An add
// would put them back while the held-out entry survives, so the slot
// has two owners, and the later restore would silently overwrite the
// power the add set. Exits from a hold go through that hold's own
// release path (unfreeze, drop, or expiry), never through a valset
// proposal.
func heldOutBy(op address) string {
	key := op.String()
	if frozenSet.Has(key) {
		return "frozen"
	}
	if autoFrozenSet.Has(key) {
		return "auto-frozen"
	}
	return ""
}

// heldOutCount is how many validators this realm is currently holding
// out of the effective valset — manually frozen plus auto-frozen.
func heldOutCount() int {
	return frozenSet.Size() + autoFrozenSet.Size()
}

// frozenPower is the voting power held out by the manual path.
//
// Split out from heldOutPower because AutoFreezeValidator needs the two
// addends separately: it uses the auto-frozen sum twice, and walking
// autoFrozenSet once instead of three times is real gas on a consensus
// path.
func frozenPower() uint64 {
	total := uint64(0)
	frozenSet.Iterate("", "", func(_ string, raw any) bool {
		total += raw.(frozenEntry).VotingPower
		return false
	})
	return total
}

// heldOutPower is the voting power of everything heldOutCount covers.
func heldOutPower() uint64 {
	return frozenPower() + autoFrozenPower()
}

// assertLivenessFloor panics unless at least freezeFloorNum/freezeFloorDen
// of the validator roster — by COUNT and by POWER — would still be live
// after removing targetPower.
//
// Both checks matter. Tendermint needs >2/3 of voting POWER to make
// progress, so under weighted voting a count-only floor is close to
// meaningless: freezing one of three validators is fine by count and
// halts the chain if that validator holds 40% of the power. The count
// floor stays as a cheap backstop against a roster shrinking to a
// handful of members.
//
// effective comes from GetValsetEffective, which ALREADY excludes
// everything this realm is holding out — hence the heldOut* addends
// that put the roster back together.
func assertLivenessFloor(effective []validators.Validator, targetPower uint64) {
	roster := len(effective) + heldOutCount()
	live := len(effective) - 1
	if live*freezeFloorDen < roster*freezeFloorNum {
		panic(ufmt.Sprintf(
			"freeze would leave %d of %d validators live, below the %d/%d floor",
			live, roster, freezeFloorNum, freezeFloorDen,
		))
	}

	livePower := uint64(0)
	for _, v := range effective {
		livePower += v.VotingPower
	}
	rosterPower := livePower + heldOutPower()
	postFreezePower := livePower - targetPower
	if postFreezePower*freezeFloorDen < rosterPower*freezeFloorNum {
		panic(ufmt.Sprintf(
			"freeze would leave %d of %d voting power live, below the %d/%d floor",
			postFreezePower, rosterPower, freezeFloorNum, freezeFloorDen,
		))
	}
}

// ---- auth -------------------------------------------------------------------

// assertFreezeAdmin panics unless the call is a direct user transaction
// signed by a freeze admin.
//
// Deliberately NOT unsafe.OriginCaller(): tx-origin identity would let
// any realm an admin happens to call freeze validators on their behalf.
// Requiring rlm.Previous() to be a user realm means the admin's own
// transaction is the immediate caller — no realm can sit in between and
// launder the authority.
// The leading `_ int` keeps this a plain helper rather than a crossing
// function, same trick as assertValopersCaller.
func assertFreezeAdmin(_ int, rlm realm) {
	// Defense-in-depth, mirroring assertValopersCaller: a stashed or
	// sibling-frame realm value would otherwise let a future call site
	// forge the designation (Class-2; see docs/resources/gno-security.md).
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}

	prev := rlm.Previous()
	if !prev.IsUser() {
		panic("freeze admins must call directly, not through realm " + prev.PkgPath())
	}
	if !freezeAdmins.Has(prev.Address().String()) {
		panic("caller is not a freeze admin: " + prev.Address().String() +
			"; use NewFreezeProposalRequest for the GovDAO path")
	}
}

// ---- shared helpers ---------------------------------------------------------

// mustValoper returns op's valoperCache entry or panics.
func mustValoper(op address) cacheEntry {
	raw := valoperCache.Get(op.String())
	if raw == nil {
		panic("unknown operator: " + op.String())
	}
	return raw.(cacheEntry)
}

// mustFrozen returns op's frozen entry or panics.
func mustFrozen(op address) frozenEntry {
	raw := frozenSet.Get(op.String())
	if raw == nil {
		panic("operator is not frozen: " + op.String())
	}
	return raw.(frozenEntry)
}

// publishValsetWithout republishes the effective valset minus one
// signing address.
func publishValsetWithout(cur realm, signingAddr address) {
	effective := sysparams.GetValsetEffective()
	set := make(map[address]validators.Validator, len(effective))
	for _, v := range effective {
		if v.Address != signingAddr {
			set[v.Address] = v
		}
	}
	publishValset(cur, set)
}

// publishValset encodes set into the sysparams wire format and
// publishes it, mirroring newValoperChangeExecutor and
// RotateValoperSigningKey.
//
// Every call site read-modify-writes against GetValsetEffective, so
// same-block writers accumulate instead of clobbering each other.
func publishValset(cur realm, set map[address]validators.Validator) {
	// The liveness floor already makes this unreachable, and the restore
	// paths only ever add power — but both other publishers guard it too:
	// an empty publish is silently swallowed by the EndBlocker, so a
	// future refactor that produced one would fail invisibly.
	if len(set) == 0 {
		panic("refusing to publish an empty validator set; consensus liveness")
	}

	entries := make([]string, 0, len(set))
	for _, v := range set {
		entries = append(entries, v.PubKey+":"+strconv.FormatUint(v.VotingPower, 10))
	}
	sort.Strings(entries)
	sysparams.SetValsetProposal(cross(cur), entries)
}
