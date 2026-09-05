package reputation

import (
	"strconv"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
)

// maxInt64 is the largest representable int64, used to guard against
// overflow when accumulating scores.
const maxInt64 = int64(1<<63 - 1)

// ledger maps target.String() -> *avl.Tree (issuerPkgPath -> *avl.Tree
// (category -> int64)). Nesting by issuer avoids any string-concatenation
// key collisions between issuer and category values, and lets Render
// enumerate per-issuer scores cheaply if needed later.
//
// totals maps target.String() -> int64, the running sum across every
// issuer and category for that target.
var (
	ledger avl.Tree
	totals avl.Tree
)

// AddPoints credits target with points under category, attributed to
// the calling realm (the "issuer"). Only callable via a crossing call
// from another realm -- caller.IsUser() covers both a direct EOA call
// and a `gnokey maketx run` ephemeral script, both of which are
// rejected here.
//
// This is attribution, not access control: nothing stops someone from
// deploying a trivial realm whose only job is to call AddPoints, so a
// high total does not by itself mean anything trustworthy. What it
// does guarantee is that every point is permanently tagged with the
// pkgpath that issued it, so a reader can see exactly which realms
// contributed to a score and discount ones they don't trust.
func AddPoints(cur realm, target address, category string, points int64) {
	caller := cur.Previous()
	if caller.IsUser() {
		panic("reputation: AddPoints can only be called by another realm, not directly by a user")
	}
	if points <= 0 {
		panic("reputation: points must be positive")
	}
	issuer := caller.PkgPath()

	issuerTree := getOrCreateSubtree(&ledger, target.String())
	categoryTree := getOrCreateSubtree(issuerTree, issuer)

	current := int64(0)
	if raw := categoryTree.Get(category); raw != nil {
		current = raw.(int64)
	}
	categoryTree.Set(category, addSafe(current, points))

	currentTotal := int64(0)
	if raw := totals.Get(target.String()); raw != nil {
		currentTotal = raw.(int64)
	}
	totals.Set(target.String(), addSafe(currentTotal, points))
}

// GetScore returns the points target has earned in category, as
// attributed by a specific issuing realm. Returns 0 if no such entry
// exists.
func GetScore(target address, category string, issuerPkgPath string) int64 {
	raw := ledger.Get(target.String())
	if raw == nil {
		return 0
	}
	issuerTree := raw.(*avl.Tree)

	raw2 := issuerTree.Get(issuerPkgPath)
	if raw2 == nil {
		return 0
	}
	categoryTree := raw2.(*avl.Tree)

	raw3 := categoryTree.Get(category)
	if raw3 == nil {
		return 0
	}
	return raw3.(int64)
}

// GetTotalScore returns target's aggregate points across every
// category and issuer.
func GetTotalScore(target address) int64 {
	raw := totals.Get(target.String())
	if raw == nil {
		return 0
	}
	return raw.(int64)
}

// getOrCreateSubtree fetches the *avl.Tree stored at key in t, creating
// and storing an empty one if none exists yet.
func getOrCreateSubtree(t *avl.Tree, key string) *avl.Tree {
	if raw := t.Get(key); raw != nil {
		return raw.(*avl.Tree)
	}
	sub := &avl.Tree{}
	t.Set(key, sub)
	return sub
}

// addSafe adds b to a, panicking on overflow rather than silently
// wrapping. Only ever called with b > 0 (AddPoints already rejects
// non-positive points), so only the positive-overflow direction needs
// checking.
func addSafe(a, b int64) int64 {
	if b > 0 && a > maxInt64-b {
		panic("reputation: score overflow")
	}
	return a + b
}

// Render shows a paginated view of total scores across all targets.
// Use ?page=<n> to navigate; page size is 20.
func Render(path string) string {
	p := pager.NewPager(&totals, 20, false)
	page := p.MustGetPageByPath(path)

	result := "# Reputation Ledger\n\n"
	result += "Points can only be added via a crossing call from another realm (see AddPoints doc comment on the attribution vs. access-control distinction).\n\n"
	result += "## Total scores\n\n"
	for _, item := range page.Items {
		result += "- " + item.Key + ": " + strconv.FormatInt(item.Value.(int64), 10) + "\n"
	}
	result += "\n" + page.Picker(path)
	return result
}
