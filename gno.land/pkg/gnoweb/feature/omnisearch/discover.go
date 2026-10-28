package omnisearch

import (
	"cmp"
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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
	if _, err := crypto.AddressFromBech32(q.Text); err == nil {
		namespaces[needle] = true
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

// addressPattern is a bech32 account address. It checks the shape only: the
// account search validates the checksum before asking the indexer.
var addressPattern = regexp.MustCompile(`^` + crypto.Bech32AddrPrefix() + `1[02-9ac-hj-np-z]{38}$`)

func isAddress(s string) bool { return addressPattern.MatchString(s) }

// sortByRelevance orders matched paths: throwaway deploys last (see
// isThrowaway), then by relevance, then the shortest path. No list of
// official namespaces: the rank follows from the paths. The omnibar ranks its
// local matches with the same rules (controller-searchbar.ts rank).
func sortByRelevance(paths []string, needle string) {
	// Each key once, not once per comparison.
	type key struct{ throwaway, score int }
	keys := make(map[string]key, len(paths))
	for _, rel := range paths {
		k := key{score: -relevance(rel, needle)}
		if isThrowaway(rel, needle) {
			k.throwaway = 1
		}
		keys[rel] = k
	}
	slices.SortFunc(paths, func(a, b string) int {
		ka, kb := keys[a], keys[b]
		return cmp.Or(
			cmp.Compare(ka.throwaway, kb.throwaway),
			cmp.Compare(ka.score, kb.score),
			cmp.Compare(len(a), len(b)),
			strings.Compare(a, b),
		)
	})
}

// relevance scores how well rel matches needle: the package's name (its last
// segment before any /vN) first, then a segment that is the needle (a
// namespace typed whole), then a match in the name, at the start of a
// segment, anywhere; a shallower path edges out a deeper one.
func relevance(rel, needle string) int {
	lower := strings.ToLower(rel)
	segs := strings.FieldsFunc(lower, func(r rune) bool { return r == '/' })
	name := gnolang.LastPathElement(lower)
	if len(segs) > 1 && name != segs[len(segs)-1] {
		segs = segs[:len(segs)-1] // a version segment is neither name nor depth
	}
	depth := len(segs)
	score := 20
	switch {
	case name == needle:
		score = 100
	case slices.Contains(segs, needle):
		score = 90
	case strings.HasPrefix(name, needle):
		score = 80
	case strings.Contains(name, needle):
		score = 60
	case slices.ContainsFunc(segs, func(s string) bool { return strings.HasPrefix(s, needle) }):
		score = 40
	}
	return score - depth
}

// isThrowaway: a deploy under a raw address namespace or a namespace with
// "test" in it, or a package segment that starts with "test" or ends in
// "_test" or "-test". Not when the needle asks for it.
func isThrowaway(rel, needle string) bool {
	ns := strings.ToLower(namespaceOf(rel))
	// A pasted address asks for that namespace; a needle inside it does not.
	if isAddress(ns) && ns != needle {
		return true
	}
	if strings.Contains(needle, "test") {
		return false
	}
	if strings.Contains(ns, "test") {
		return true
	}
	// Not "latest" or "attestation".
	for _, seg := range strings.Split(strings.ToLower(rel), "/") {
		if strings.HasPrefix(seg, "test") || strings.HasSuffix(seg, "_test") || strings.HasSuffix(seg, "-test") {
			return true
		}
	}
	return false
}
