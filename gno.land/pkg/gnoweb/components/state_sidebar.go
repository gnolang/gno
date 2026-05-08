package components

import (
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// reAnchorSafe matches characters that don't need escaping inside an HTML
// id / URL-fragment. Used to derive anchors from declaration names.
var reAnchorSafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// ShortenOID returns id's trailing `:N` when its 40-char hashlet
// matches ref's, otherwise the full id. Used by both the sidebar
// builder and the `oidShort` template func — single source of truth.
func ShortenOID(id, ref string) string {
	i, j := strings.IndexByte(id, ':'), strings.IndexByte(ref, ':')
	if i > 0 && j > 0 && id[:i] == ref[:j] {
		return id[i:]
	}
	return id
}

// truncMid shortens a long string to "<head>…<tail>". Returns the original
// when it's already short enough that truncation would save nothing.
// Used to make 40-char hashes / OIDs scannable in narrow chips while
// keeping the full text reachable via the surrounding `title` attribute.
func truncMid(s string, head, tail int) string {
	if head < 0 {
		head = 0
	}
	if tail < 0 {
		tail = 0
	}
	if len(s) <= head+tail+1 {
		return s
	}
	return s[:head] + "…" + s[len(s)-tail:]
}

// TruncOID truncates an ObjectID's hashlet to `head…tail` while
// preserving the `:N` suffix (the part that distinguishes objects in
// the same realm). Plain non-OID strings get a bare middle truncation.
func TruncOID(id string, head, tail int) string {
	if i := strings.IndexByte(id, ':'); i > 0 {
		return truncMid(id[:i], head, tail) + id[i:]
	}
	return truncMid(id, head, tail)
}

// BuildPackageSidebar assembles the aside content for a top-level state
// page (e.g. /r/foo$state). Two sections: "Realm" (path + kind) and
// "Stats" (declaration count) — followed by the TOC of decls.
func BuildPackageSidebar(pkgPath string, nodes []StateNode) *StateSidebar {
	if len(nodes) == 0 {
		return nil
	}
	heading := "Top-level declarations"
	if !strings.HasPrefix(pkgPath, "/r/") {
		heading = "Package declarations"
	}
	kindLabel := PkgKindLabel(pkgPath)
	return &StateSidebar{
		Heading: heading,
		TOC:     buildTOC(nodes),
		Meta: []StateMetaEntry{
			{Section: kindLabel, Label: "Path", Value: pkgPath, Mono: true},
			{Section: "Stats", Label: "Declarations", Value: fmt.Sprintf("%d", len(nodes)), Inline: true},
		},
	}
}

// BuildObjectSidebar assembles the aside content for a per-object state
// page (e.g. /r/foo$state&oid=X). Meta is grouped into Identity (Realm,
// OID, Type), Lineage (Owner) and Storage (Size, Refs, Mod, Hash) so
// users can scan the audit info quickly. Long IDs/hashes are mono +
// truncated; short numeric values render inline. */
func BuildObjectSidebar(pkgPath, oid, typeID string, height int64, info StateObjectInfoView, nodes []StateNode) *StateSidebar {
	meta := []StateMetaEntry{
		{Section: "Identity", Label: "Realm", Value: pkgPath, Href: RealmStateHref(pkgPath)},
		{Label: "Object ID", Value: oid, Mono: true},
	}
	if typeID != "" {
		meta = append(meta, StateMetaEntry{Label: "Type", Value: typeID, Mono: true})
	}
	if info.OwnerID != "" {
		// Owner navigation: clicking the owner takes you to its own state
		// page at the SAME height so time-travel holds across the hop.
		// When the owner shares the queried object's hashlet (same realm),
		// only show the trailing `:N` so the row doesn't repeat ~99% of
		// the OID just above.
		meta = append(meta, StateMetaEntry{
			Section: "Lineage", Label: "Owner", Value: ShortenOID(info.OwnerID, oid),
			Href: stateObjectHref(pkgPath, info.OwnerID, "", height),
			Mono: true,
		})
	}
	// Storage section — inline pairs for short numerics, block for hash.
	storageEntries := []StateMetaEntry{}
	if info.LastObjectSize != "" {
		storageEntries = append(storageEntries, StateMetaEntry{Label: "Size", Value: info.LastObjectSize + " B", Inline: true})
	}
	if info.RefCount != "" {
		storageEntries = append(storageEntries, StateMetaEntry{Label: "Refs", Value: info.RefCount, Inline: true})
	}
	if info.ModTime != "" {
		storageEntries = append(storageEntries, StateMetaEntry{Label: "Modified", Value: "#" + info.ModTime, Inline: true})
	}
	if info.Hash != "" {
		storageEntries = append(storageEntries, StateMetaEntry{Label: "Hash", Value: info.Hash, Mono: true})
	}
	if len(storageEntries) > 0 {
		storageEntries[0].Section = "Storage"
		meta = append(meta, storageEntries...)
	}
	if info.IsEscaped {
		meta = append(meta, StateMetaEntry{Section: "Status", Label: "Escaped", Value: "yes", Inline: true})
	}

	return &StateSidebar{
		Heading: "Fields",
		TOC:     buildTOC(nodes),
		Meta:    meta,
	}
}

// buildTOC turns a list of top-level StateNodes into TOC entries. The
// anchor is a lowercased, dash-safe slug that the template stamps as
// `id="<anchor>"` on the matching row. Mutates each node's Anchor field
// so the template can read it back when rendering the row.
func buildTOC(nodes []StateNode) []StateTOCEntry {
	entries := make([]StateTOCEntry, 0, len(nodes))
	seen := make(map[string]int)
	for i := range nodes {
		n := &nodes[i]
		base := stateAnchorOf(n.Name)
		anchor := base
		if seen[base] > 0 {
			// De-duplicate identical labels (e.g. positional indices) so
			// every TOC entry has a unique scroll target.
			anchor = fmt.Sprintf("%s-%d", base, seen[base])
		}
		seen[base]++
		n.Anchor = anchor
		entries = append(entries, StateTOCEntry{
			Label:  n.Name,
			Anchor: anchor,
			Kind:   n.Kind,
			Type:   n.Type,
		})
	}
	return entries
}

// stateAnchorOf turns a node name into a fragment-safe anchor identifier.
// The output is always non-empty (falls back to "decl" for empty input).
func stateAnchorOf(name string) string {
	clean := reAnchorSafe.ReplaceAllString(strings.TrimSpace(name), "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return "decl"
	}
	return "state-" + strings.ToLower(clean)
}

// RealmStateHref returns the URL of a package's top-level state page
// (`/r/foo$state`). Single source for any caller that needs a "back to
// realm" link — sidebar Identity row, object-page breadcrumb — so the
// `$state` query syntax lives in one place.
func RealmStateHref(pkgPath string) template.URL {
	u := weburl.GnoURL{Path: pkgPath, WebQuery: url.Values{"state": {""}}}
	return template.URL(u.EncodeWebURL())
}

// PkgKindLabel returns the human-readable label for a package path:
// "Realm" for paths under `/r/`, "Package" otherwise. Used by both
// the sidebar Identity row and the state-page title — single source
// so a future "Library" or "App" classification lives in one place.
func PkgKindLabel(pkgPath string) string {
	if strings.HasPrefix(pkgPath, "/r/") {
		return "Realm"
	}
	return "Package"
}
