package components

import (
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// reAnchorSafe matches characters unsafe inside an HTML id / URL fragment.
var reAnchorSafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// ShortenOID returns id's trailing `:N` when its hashlet matches ref's,
// otherwise the full id.
func ShortenOID(id, ref string) string {
	i, j := strings.IndexByte(id, ':'), strings.IndexByte(ref, ':')
	if i > 0 && j > 0 && id[:i] == ref[:j] {
		return id[i:]
	}
	return id
}

// truncMid shortens s to "<head>…<tail>", or returns it unchanged when short.
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

// TruncOID truncates an ObjectID's hashlet while preserving the `:N` suffix.
func TruncOID(id string, head, tail int) string {
	if i := strings.IndexByte(id, ':'); i > 0 {
		return truncMid(id[:i], head, tail) + id[i:]
	}
	return truncMid(id, head, tail)
}

// BuildPackageSidebar assembles the aside for a top-level state page.
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
			{Section: kindLabel, Label: "Path", Value: pkgPath},
			{Section: "Stats", Label: "Declarations", Value: fmt.Sprintf("%d", len(nodes)), Inline: true},
		},
	}
}

// BuildObjectSidebar assembles the aside for a per-object state page,
// grouping meta into Identity, Lineage, and Storage sections.
func BuildObjectSidebar(pkgPath, oid, typeID string, height int64, info StateObjectInfoView, nodes []StateNode) *StateSidebar {
	meta := []StateMetaEntry{
		{Section: "Identity", Label: "Realm", Value: pkgPath, Href: RealmStateHref(pkgPath)},
		{Label: "Object ID", Value: oid, Mono: true},
	}
	if typeID != "" {
		meta = append(meta, StateMetaEntry{Label: "Type", Value: typeID, Mono: true})
	}
	if info.OwnerID != "" {
		// Owner link preserves height so time-travel holds across the hop.
		meta = append(meta, StateMetaEntry{
			Section: "Lineage", Label: "Owner", Value: ShortenOID(info.OwnerID, oid),
			Href: stateObjectHref(pkgPath, info.OwnerID, "", height),
			Mono: true,
		})
	}
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

	fields := "fields"
	if len(nodes) == 1 {
		fields = "field"
	}
	return &StateSidebar{
		Heading: fmt.Sprintf("%d %s", len(nodes), fields),
		TOC:     buildTOC(nodes),
		Meta:    meta,
	}
}

// buildTOC builds TOC entries from top-level StateNodes, mutating each
// node's Anchor so the template can stamp `id="<anchor>"` on the row.
func buildTOC(nodes []StateNode) []StateTOCEntry {
	entries := make([]StateTOCEntry, 0, len(nodes))
	seen := make(map[string]int)
	for i := range nodes {
		n := &nodes[i]
		base := stateAnchorOf(n.Name)
		anchor := base
		if seen[base] > 0 {
			// De-duplicate identical labels so every entry has a unique target.
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

// stateAnchorOf turns a node name into a fragment-safe anchor (never empty).
func stateAnchorOf(name string) string {
	clean := reAnchorSafe.ReplaceAllString(strings.TrimSpace(name), "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return "decl"
	}
	return "state-" + strings.ToLower(clean)
}

// RealmStateHref returns the URL of a package's top-level state page (`/r/foo$state`).
func RealmStateHref(pkgPath string) template.URL {
	u := weburl.GnoURL{Path: pkgPath, WebQuery: url.Values{"state": {""}}}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
}

// PkgKindLabel returns "Realm" for `/r/` paths, "Package" otherwise.
func PkgKindLabel(pkgPath string) string {
	if strings.HasPrefix(pkgPath, "/r/") {
		return "Realm"
	}
	return "Package"
}
