// Package teams holds, for each namespace, the set of addresses allowed to
// deploy under it, and whether that namespace still honours the single
// registered-name address r/sys/users resolves.
//
// Authority to write a namespace's entry is not granted by an argument, a
// whitelist or a controller. It is derived from the calling realm's own package
// path: only gno.land/r/<ns>/home may write <ns>. There is therefore no list to
// curate and no controller address to compromise.
//
// The same gate shape runs in r/sys/params (assertDelegate) and
// r/sys/validators/v0 (assertValopersCaller). The three steps below are copied
// from them deliberately; see assertHome for why each one is load bearing.
//
// This realm never calls back into r/<ns>/home. Membership is pushed in as
// arguments. Besides the fact that gno has no dynamic dispatch by package path,
// a callback would make home see this realm as its previous realm, which is the
// confused-deputy shape r/sys/validators/v0 documents having eliminated.
package teams

import (
	"chain"
	"chain/runtime"
	"strings"

	"gno.land/p/nt/addrset/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

// homeSuffix anchors the caller to the one realm named for the namespace.
const homeSuffix = "/home"

// realmPrefix is the only package-path prefix a caller may present.
//
// Derived from the chain's own domain rather than hardcoded: ChainDomain is a
// vm parameter, and a hardcoded "gno.land/" would make this realm refuse every
// caller on a chain configured with any other domain. That failure is silent
// and reads as a dead realm rather than a misconfiguration.
//
// Requiring /r/ is also what excludes a `maketx run` realm, which presents
// <domain>/e/<addr>/run: a security boundary, not a naming convention.
func realmPrefix() string {
	return runtime.ChainDomain() + "/r/"
}

// Events emitted on every state change, so an indexer can follow a namespace's
// authority without replaying the tree.
const (
	MembersSetEvent   = "TeamMembersSet"
	MemberAddedEvent  = "TeamMemberAdded"
	MemberRemovedEvent = "TeamMemberRemoved"
	RetiredEvent      = "TeamNameAuthorityRetired"
)

// team is one namespace's authority record.
type team struct {
	members addrset.Set

	// retired reports that this namespace no longer honours the address
	// r/sys/users resolves for its name. Once true the team is the only way
	// to deploy under the namespace. One-way by design: see Retire.
	retired bool
}

// teams maps namespace -> *team.
var teams = avl.NewTree()

// assertHome authorizes the caller and returns the namespace it owns.
//
// Three steps, none of them optional, each one load bearing for a reason that
// is not obvious from reading it:
//
//  1. IsCurrent. The helper trusts its rlm input, so a caller threading a
//     stashed or sibling-frame realm value would otherwise present someone
//     else's designation. docs/resources/gno-security.md names the class.
//
//  2. The empty check, before anything parses the path. A direct user MsgCall
//     presents the empty package path. This realm derives the namespace FROM
//     the caller rather than comparing it to a stored value, so an unguarded
//     empty path does not fail a comparison, it parses into a namespace. It has
//     to be refused explicitly.
//
//  3. Exact structure, never a prefix. cur.Sub(subpath) presents the
//     synthesized identity "host#subpath", and a namespace may hold nested
//     realms, so gno.land/r/<ns>/home#x and gno.land/r/<ns>/x/home must both be
//     refused. Sub-realms of home do not inherit the namespace.
//
// Takes rlm in the non-crossing dispatch position, matching assertDelegate in
// r/sys/params, so a caller cannot route through this helper and have it cross
// on their behalf.
func assertHome(_ int, rlm realm) string {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	caller := rlm.Previous().PkgPath()
	if caller == "" {
		panic("unauthorized: no calling realm; r/<ns>/home must call this")
	}
	ns, ok := namespaceOfHome(caller)
	if !ok {
		panic("unauthorized: caller is not a <ns>/home realm: " + caller)
	}
	return ns
}

// namespaceOfHome returns the namespace of a gno.land/r/<ns>/home package path.
//
// It accepts exactly that shape: the realm prefix, one path segment, then
// /home. Anything else, including a deeper path, a sub-realm identity and the
// empty namespace, is refused.
func namespaceOfHome(pkgPath string) (string, bool) {
	prefix := realmPrefix()
	if !strings.HasPrefix(pkgPath, prefix) {
		return "", false
	}
	rest := pkgPath[len(prefix):]
	if !strings.HasSuffix(rest, homeSuffix) {
		return "", false
	}
	ns := rest[:len(rest)-len(homeSuffix)]
	if ns == "" {
		return "", false
	}
	// A namespace is one segment. "/" catches r/<ns>/x/home, "#" catches the
	// sub-realm identity cur.Sub() synthesizes.
	if strings.ContainsAny(ns, "/#") {
		return "", false
	}
	return ns, true
}

// getOrCreate returns the namespace's record, creating it on first write.
func getOrCreate(ns string) *team {
	if v, ok := teams.Get(ns).(*team); ok {
		return v
	}
	t := &team{}
	teams.Set(ns, t)
	return t
}

func get(ns string) (*team, bool) {
	v, ok := teams.Get(ns).(*team)
	return v, ok
}

// SeedMembers sets the member list only if this namespace has no record yet.
//
// This is the function init() should call. init() re-runs on every redeploy,
// over the prior realm's state, so an init() calling SetMembers would turn a
// routine redeploy of home into an undeclared reset of the team. Seeding is
// idempotent, so the same init() is safe to re-run forever.
//
// Returns true if it wrote.
func SeedMembers(cur realm, members []address) bool {
	ns := assertHome(0, cur)
	if _, exists := get(ns); exists {
		return false
	}
	setMembers(ns, members)
	return true
}

// SetMembers replaces this namespace's member list.
//
// Call it from an exported helper on home, never from init(): see SeedMembers.
func SetMembers(cur realm, members []address) {
	ns := assertHome(0, cur)
	t := getOrCreate(ns)
	if t.retired && len(members) == 0 {
		panic("refused: emptying the team of a retired namespace would strand it")
	}
	setMembers(ns, members)
}

// setMembers replaces the member set by diffing against the live one, so a
// no-op write costs nothing and a one-address change does not orphan the whole
// backing tree in persisted state.
func setMembers(ns string, members []address) {
	t := getOrCreate(ns)
	if len(members) > MaxMembers {
		panic(ufmt.Sprintf("refused: %d members exceeds the %d cap", len(members), MaxMembers))
	}
	want := map[string]bool{}
	for _, m := range members {
		if !m.IsValid() {
			panic("invalid member address: " + m.String())
		}
		want[string(m)] = true
	}
	// Collect first: the backing B+ tree must not be mutated from inside its
	// own iteration callback.
	var drop []address
	t.members.IterateByOffset(0, t.members.Size(), func(a address) bool {
		if !want[string(a)] {
			drop = append(drop, a)
		}
		return false
	})
	for _, a := range drop {
		t.members.Remove(a)
	}
	for _, m := range members {
		t.members.Add(m)
	}
	chain.Emit(MembersSetEvent, "namespace", ns, "size", ufmt.Sprintf("%d", t.members.Size()))
}

// AddMember adds one address to this namespace's team. Returns false if it was
// already a member.
func AddMember(cur realm, addr address) bool {
	ns := assertHome(0, cur)
	if !addr.IsValid() {
		panic("invalid member address: " + addr.String())
	}
	t := getOrCreate(ns)
	if t.members.Size() >= MaxMembers && !t.members.Has(addr) {
		panic(ufmt.Sprintf("refused: team is at the %d member cap", MaxMembers))
	}
	if !t.members.Add(addr) {
		return false
	}
	chain.Emit(MemberAddedEvent, "namespace", ns, "address", addr.String())
	return true
}

// RemoveMember drops one address. Returns false if it was not a member.
//
// Removing the last member of a retired namespace is refused: with the
// registered-name path already retired, nobody could deploy under it again and
// nobody could re-form the team, because forming it requires deploying home.
func RemoveMember(cur realm, addr address) bool {
	ns := assertHome(0, cur)
	t, ok := get(ns)
	if !ok {
		return false
	}
	if t.retired && t.members.Size() == 1 && t.members.Has(addr) {
		panic("refused: removing the last member of a retired namespace would strand it")
	}
	if !t.members.Remove(addr) {
		return false
	}
	chain.Emit(MemberRemovedEvent, "namespace", ns, "address", addr.String())
	return true
}

// Retire stops this namespace honouring the single address r/sys/users resolves
// for its name. After it, only the team may deploy under the namespace.
//
// This is the "burn" step of a migration: the original key publishes home with
// the new owners, moves its coins, then retires its own direct authority.
//
// One-way, and refused while the team is empty, because the combination is a
// namespace nobody can ever deploy under again.
func Retire(cur realm) {
	ns := assertHome(0, cur)
	t := getOrCreate(ns)
	if t.members.Size() == 0 {
		panic("refused: retiring with an empty team would strand the namespace")
	}
	if t.retired {
		return
	}
	t.retired = true
	chain.Emit(RetiredEvent, "namespace", ns)
}

// IsMember reports whether addr may deploy under ns by team membership.
// This is the read r/sys/names' third authorization path makes.
func IsMember(ns string, addr address) bool {
	t, ok := get(ns)
	if !ok {
		return false
	}
	return t.members.Has(addr)
}

// IsRetired reports whether ns has given up the registered-name path.
// r/sys/names must skip its registered-name check when this is true.
func IsRetired(ns string) bool {
	t, ok := get(ns)
	if !ok {
		return false
	}
	return t.retired
}

// HasTeam reports whether ns has a record at all.
func HasTeam(ns string) bool {
	_, ok := get(ns)
	return ok
}

// MaxMembers caps a team.
//
// The cap is not about storage, which a B+ tree handles, but about every read
// that walks the set. Without it one namespace could make MembersOf and Render
// unbounded work for anyone who calls them, including an r/sys realm.
const MaxMembers = 128

// MembersOf returns one page of ns's members in sorted order, or nil.
//
// Paginated rather than returning the whole set: an unbounded return is
// unbounded gas for every caller, and this is a read any realm may make.
// count is clamped to MaxMembers.
func MembersOf(ns string, offset, count int) []address {
	t, ok := get(ns)
	if !ok || count <= 0 || offset < 0 {
		return nil
	}
	if count > MaxMembers {
		count = MaxMembers
	}
	out := make([]address, 0, count)
	t.members.IterateByOffset(offset, count, func(a address) bool {
		out = append(out, a)
		return false
	})
	return out
}

// MemberCount returns how many addresses ns's team holds.
func MemberCount(ns string) int {
	t, ok := get(ns)
	if !ok {
		return 0
	}
	return t.members.Size()
}


