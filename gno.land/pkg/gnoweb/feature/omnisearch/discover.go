package omnisearch

import (
	"context"
	"sort"
	"strings"
)

// maxDiscoverResults caps each discovery group. The omnibar shows what is
// relevant, not everything the chain holds.
const maxDiscoverResults = 10

// discover answers a query that names no selector. One directory listing,
// coalesced with every other caller; the users group is derived from the
// paths already fetched, so it is free.
func (h *Handler) discover(ctx context.Context, q *Query) []Group {
	needle := strings.ToLower(q.Text)
	author, _ := q.Get(FilterAuthor)
	// `author:` bypassing the floor meant a bare `author:` — no value at all
	// — bought a full directory listing and returned nothing.
	if len(needle) < MinTermLen && len(author) < MinTermLen {
		return nil
	}

	realms, packages, truncated, err := h.deps.Directory.Paths(ctx)
	if err != nil {
		return []Group{{Label: "Realms", Source: SourceChain, Err: err}}
	}

	kinds := []struct {
		label string
		is    string
		paths []string
	}{
		{"Realms", "realm", realms},
		{"Packages", "package", packages},
	}
	want, hasIs := q.Get(FilterIs)

	var (
		groups     []Group
		namespaces = map[string]bool{}
	)
	for _, k := range kinds {
		if hasIs && !strings.EqualFold(want, k.is) {
			continue
		}

		g := Group{Label: k.label, Source: SourceChain}
		for _, p := range k.paths {
			rel := strings.TrimPrefix(p, h.deps.Domain)
			// Needle first: it rejects most paths and costs nothing.
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
			if len(g.Results) >= maxDiscoverResults {
				continue
			}
			g.Results = append(g.Results, Result{
				Title:  rel,
				Detail: ns,
				Href:   safePathHref(rel),
				Tags:   []string{k.is},
			})
		}
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
		for i := range groups {
			groups[i].Truncated = true
		}
	}
	return groups
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
