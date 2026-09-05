package docs

import (
	"io/fs"
	"regexp"
	"strings"
	"sync"
)

// SidebarSection is one top-level category in the docs navigation (a `## …`
// heading in README.md, e.g. "Use Gno.land", "Build on Gno.land", "References").
type SidebarSection struct {
	Title string
	Items []SidebarItem
}

// SidebarItem is a single entry under a section. Internal items map to an
// embedded .md file (Href is the embed-relative path, e.g.
// "builders/getting-started.md"); external items have an absolute URL.
type SidebarItem struct {
	Title    string
	Href     string
	Summary  string
	External bool
}

// sidebarItemRE matches a single bullet item in a section list:
//
//   - [Title](href) - optional summary text
//
// Submatches: 1 = title, 2 = href, 3 = summary (may be empty). Backticks
// and other inline markdown inside the title are kept verbatim so that the
// renderer can decide how to display them.
var sidebarItemRE = regexp.MustCompile(`^\s*-\s+\[(.+?)\]\(([^)]+)\)(?:\s*-\s*(.*))?$`)

// sidebarSectionRE matches a level-2 heading: "## Use Gno.land".
var sidebarSectionRE = regexp.MustCompile(`^##\s+(.+?)\s*$`)

var (
	sidebarOnce  sync.Once
	sidebarValue []SidebarSection
)

// Sidebar returns the navigation tree parsed from README.md. The parse runs
// once on first call and is cached. Returning a fresh slice on every call
// would also work given the small size, but caching keeps the cost flat
// regardless of how many pages render.
func Sidebar() []SidebarSection {
	sidebarOnce.Do(func() {
		src, err := fs.ReadFile(FS(), "README.md")
		if err != nil {
			// README.md is part of the embed; a missing file would be a
			// build-time problem, not a runtime one. Return empty rather
			// than panic so a degraded gnoweb still serves docs without
			// a sidebar.
			return
		}
		sidebarValue = parseSidebar(src)
	})
	return sidebarValue
}

// parseSidebar walks the README line by line collecting sections and items.
// Anything not matching a section heading or a bullet pattern is ignored
// (intro prose, blank lines, etc.).
func parseSidebar(src []byte) []SidebarSection {
	var (
		sections []SidebarSection
		cur      *SidebarSection
		inFence  bool
	)
	for line := range strings.SplitSeq(string(src), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if m := sidebarSectionRE.FindStringSubmatch(line); m != nil {
			sections = append(sections, SidebarSection{Title: strings.TrimSpace(m[1])})
			cur = &sections[len(sections)-1]
			continue
		}
		if cur == nil {
			continue
		}
		if m := sidebarItemRE.FindStringSubmatch(line); m != nil {
			href := strings.TrimSpace(m[2])
			cur.Items = append(cur.Items, SidebarItem{
				Title:    strings.TrimSpace(m[1]),
				Href:     href,
				Summary:  strings.TrimSpace(m[3]),
				External: isExternalHref(href),
			})
		}
	}
	return sections
}

func isExternalHref(href string) bool {
	return strings.HasPrefix(href, "http://") ||
		strings.HasPrefix(href, "https://") ||
		strings.HasPrefix(href, "//")
}
