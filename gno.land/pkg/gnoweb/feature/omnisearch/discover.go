package omnisearch

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// maxDiscoverResults caps each discovery group. The omnibar shows what is
// relevant, not everything the chain holds.
const maxDiscoverResults = 10

// discover answers a query that names no selector. One directory listing,
// coalesced with every other caller; the users group is derived from the
// paths already fetched, so it is free.
//
// Only an explicit `in:` narrows it (see discoveryScope).
func (h *Handler) discover(ctx context.Context, q *Query) []Group {
	needle := strings.ToLower(q.Text)
	author, _ := q.Get(FilterAuthor)
	scope := q.discoveryScope()
	// `author:` bypassing the floor meant a bare `author:` — no value at all
	// — bought a full directory listing and returned nothing. A scope is
	// narrow enough on its own.
	if len(needle) < MinTermLen && len(author) < MinTermLen && scope == "" {
		return refused(fmt.Sprintf("type at least %d characters, or a qualifier such as author:", MinTermLen))
	}
	want, hasIs := q.Get(FilterIs)
	if hasIs && !strings.EqualFold(want, "realm") && !strings.EqualFold(want, "package") {
		return refused(fmt.Sprintf("is:%s is not a kind: use is:realm or is:package", want))
	}

	realms, packages, truncated, err := h.deps.Directory.Paths(ctx)
	if err != nil {
		h.deps.Logger.Warn("omnisearch: path listing failed", "error", err)
		return []Group{{Label: "Realms", Source: SourceChain, Err: publicError(err, SourceChain)}}
	}

	kinds := []struct {
		label string
		is    string
		paths []string
	}{
		{"Realms", "realm", realms},
		{"Packages", "package", packages},
	}
	var (
		groups     []Group
		namespaces = map[string]bool{}
	)
	for _, k := range kinds {
		if hasIs && !strings.EqualFold(want, k.is) {
			continue
		}

		g := Group{Label: k.label, Source: SourceChain}
		var hits []string
		for _, p := range k.paths {
			rel := strings.TrimPrefix(p, h.deps.Domain)
			// Scope first, then the needle: both reject most paths, and the
			// scope check does not allocate.
			if !inScope(rel, scope) {
				continue
			}
			if needle != "" && !strings.Contains(strings.ToLower(rel), needle) {
				continue
			}
			ns := namespaceOf(rel)
			if author != "" && !strings.EqualFold(ns, author) {
				continue
			}
			// Recorded before the cap, so the users group reflects every
			// match rather than the first ten.
			if ns != "" {
				namespaces[ns] = true
			}
			hits = append(hits, rel)
		}
		// Ranked before the cap: the listing is alphabetical, which put a
		// test deploy under an address ahead of the package everyone uses.
		sortByRelevance(hits, needle)
		matched := len(hits)
		for _, rel := range hits[:min(matched, maxDiscoverResults)] {
			g.Results = append(g.Results, Result{
				Title:  rel,
				Detail: namespaceOf(rel),
				Href:   safePathHref(rel),
				Tags:   []string{k.is},
			})
		}
		markCapped(&g, matched)
		// An empty group is not an answer. A failed one still renders:
		// "could not ask" is not "nothing matched".
		if len(g.Results) > 0 || g.Err != nil {
			groups = append(groups, g)
		}
	}

	// A pasted address names a user even with nothing deployed under it.
	if isAddress(q.Text) {
		namespaces[strings.ToLower(strings.TrimSpace(q.Text))] = true
	}
	if g := usersGroup(namespaces); len(g.Results) > 0 {
		groups = append(groups, g)
	}
	if truncated {
		// The node always drops the same lexicographic tail, so a namespace
		// late in the alphabet would otherwise look like it does not exist.
		// With no match in the part listed, the notice is the answer, and
		// needs a group to carry it.
		if len(groups) == 0 {
			groups = append(groups, Group{Label: "Paths", Source: SourceChain})
		}
		for i := range groups {
			groups[i].Truncated = true
			groups[i].Notice = strings.TrimSpace(groups[i].Notice + " " + listingCapNotice)
		}
	}
	return groups
}

// listingCapNotice explains a node-capped listing.
const listingCapNotice = "Showing part of the chain: the node caps this listing, so more may exist."

// markCapped says how many matched when a group shows fewer: the count in
// its header reads the rows kept, not the matches.
func markCapped(g *Group, matched int) {
	if matched > len(g.Results) {
		g.Truncated = true
		g.Notice = fmt.Sprintf("Showing %d of %d matches.", len(g.Results), matched)
	}
}

// refused is the answer to a query discovery will not run: why, rather than
// "Nothing matched.".
func refused(why string) []Group {
	return []Group{{Label: "Search", Source: SourceChain, Err: inputError(why)}}
}

// discoveryScope is the package a discovery search is narrowed to: the one
// named by `in:`, never the page the search was typed on. The omnibar sends
// every query from the page path, so taking the page as a scope would turn
// `author:demo` or a bare word on a realm page into a search of that realm
// alone.
func (q *Query) discoveryScope() string {
	if _, ok := q.Get(FilterIn); !ok {
		return ""
	}
	return q.PkgPath
}

// inScope reports whether rel is the scoped package or sits under it. No
// scope admits every path.
func inScope(rel, scope string) bool {
	return scope == "" || rel == scope || strings.HasPrefix(rel, scope+"/")
}

// usersGroup is derived, not fetched: every namespace came from a path the
// chain already returned.
func usersGroup(namespaces map[string]bool) Group {
	names := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		names = append(names, ns)
	}
	sort.Strings(names)

	g := Group{Label: "Users", Source: SourceChain}
	for _, ns := range names {
		if len(g.Results) >= maxDiscoverResults {
			break
		}
		g.Results = append(g.Results, Result{Title: ns, Href: safeUserHref(ns)})
	}
	markCapped(&g, len(names))
	return g
}

// namespaceOf returns the owner segment of a gnoweb path:
// "/r/demo/boards" -> "demo". Local rather than weburl.Namespace() because
// this runs over every listed path and that one re-validates with a regexp.
func namespaceOf(rel string) string {
	_, rest, ok := strings.Cut(strings.TrimPrefix(rel, "/"), "/")
	if !ok {
		return ""
	}
	ns, _, _ := strings.Cut(rest, "/")
	return ns
}

// matchesAuthor accepts either the namespace or the deploying address: a
// reader may know either.
func matchesAuthor(chainPath, signer, author, domain string) bool {
	if author == "" {
		return true
	}
	if strings.EqualFold(signer, author) {
		return true
	}
	return strings.EqualFold(namespaceOf(strings.TrimPrefix(chainPath, domain)), author)
}

// isAddress reports whether s, trimmed, is a bech32 account address.
func isAddress(s string) bool {
	_, err := crypto.AddressFromBech32(strings.TrimSpace(s))
	return err == nil
}

// sortByRelevance orders matched paths: a path named after the needle first,
// a throwaway deploy last, then the shortest. A throwaway is one under a raw
// address namespace or with a "test" segment, unless the needle asks for it.
// No list of official namespaces: what ranks first follows from the paths.
func sortByRelevance(paths []string, needle string) {
	// Each key once: a throwaway check decodes a bech32 namespace.
	ranks := make(map[string][3]int, len(paths))
	for _, rel := range paths {
		r := [3]int{0, 0, len(rel)}
		if isThrowaway(rel, needle) {
			r[0] = 1
		}
		if !namedAfter(rel, needle) {
			r[1] = 1
		}
		ranks[rel] = r
	}
	slices.SortFunc(paths, func(a, b string) int {
		ra, rb := ranks[a], ranks[b]
		if c := slices.Compare(ra[:], rb[:]); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	})
}

// namedAfter reports whether the package's own name, its last segment before
// any version (`/v0`), is the needle.
func namedAfter(rel, needle string) bool {
	segs := strings.Split(strings.Trim(rel, "/"), "/")
	last := segs[len(segs)-1]
	if isVersion(last) && len(segs) > 1 {
		last = segs[len(segs)-2]
	}
	return needle != "" && strings.EqualFold(last, needle)
}

// isVersion matches a version segment such as v0 or v12.
func isVersion(seg string) bool {
	if len(seg) < 2 || seg[0] != 'v' {
		return false
	}
	for _, c := range seg[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isThrowaway: a deploy under a raw address namespace, or a test package.
func isThrowaway(rel, needle string) bool {
	ns := namespaceOf(rel)
	if isAddress(ns) && !strings.Contains(ns, needle) {
		return true
	}
	if strings.Contains(needle, "test") {
		return false
	}
	// A namespace is a name someone picked: "moultest123" is a test one.
	if strings.Contains(strings.ToLower(ns), "test") {
		return true
	}
	// A package segment: "test", "testing", "foo_test", not "latest" or
	// "attestation".
	for _, seg := range strings.Split(strings.ToLower(rel), "/") {
		if strings.HasPrefix(seg, "test") || strings.HasSuffix(seg, "_test") || strings.HasSuffix(seg, "-test") {
			return true
		}
	}
	return false
}
