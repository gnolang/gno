package omnisearch

import (
	"context"
	"fmt"
	"sort"
	"strings"
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
		matched := 0
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
			if matched++; len(g.Results) >= maxDiscoverResults {
				continue
			}
			g.Results = append(g.Results, Result{
				Title:  rel,
				Detail: ns,
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
