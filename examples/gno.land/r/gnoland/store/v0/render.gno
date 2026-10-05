package store

import (
	"chain/runtime"
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/moul/md/v0"
)

// storeURL is this realm's gnoweb path, wherever it is deployed.
var storeURL string

func init(cur realm) {
	storeURL = strings.TrimPrefix(cur.PkgPath(), runtime.ChainDomain())
}

// Render serves the JSON API under api/ and a markdown store otherwise: ""
// (home), "build", "c/<category>[?page=N]" and the lists: "top", "latest",
// "updated" (paged with ?page=N) and "trending". gnoweb renders these paths
// richer and falls back to this markdown. A listing has no page here: its
// own realm or package page is its detail page.
func Render(path string) string {
	path, query, _ := strings.Cut(path, "?")
	if api, ok := strings.CutPrefix(path, "api/"); ok {
		return renderAPI(api)
	}
	h := runtime.ChainHeight()
	switch path {
	case "":
		return renderHome(h)
	case "build":
		return renderBuild(h)
	}
	if l, ok := findList(path, h); ok {
		return renderPaged(l, storeURL+":"+l.Key, query)
	}
	if key, ok := strings.CutPrefix(path, "c/"); ok {
		if c, found := findCategory(key); found {
			return renderPaged(list{Title: c.Label, Index: categoryMembers(c.Key)}, storeURL+":c/"+c.Key, query)
		}
	}
	return md.H1("Not found") + "\n" + md.Link("Back to Explore", storeURL) + "\n"
}

func renderHome(h int64) string {
	var b strings.Builder
	b.WriteString(md.H1("Explore") + "\n")
	b.WriteString("Apps you can read. Everything here runs on gno.land and its code is public: try an app, or read how it works first. " +
		md.Link("Build", storeURL+":build") + " lists packages and services.\n\n")
	b.WriteString(md.H2("Categories") + "\n")
	for _, c := range categories {
		b.WriteString(md.BulletItem(md.Link(c.Label, storeURL+":c/"+c.Key) + " (" + strconv.Itoa(appCount(c.Key)) + ")"))
	}
	if l := picked(h); l != nil {
		b.WriteString("\n" + md.H2("Picked by GovDAO") + "\n")
		writeItems(&b, []*listing{l})
	}
	writeMarkdownShelves(&b, homeShelves(h))
	return b.String()
}

func renderBuild(h int64) string {
	var b strings.Builder
	b.WriteString(md.H1("Build") + "\nPackages and services to build on.\n")
	writeMarkdownShelves(&b, buildShelves(h))
	b.WriteString("\n" + md.Link("Back to Explore", storeURL) + "\n")
	return b.String()
}

func writeMarkdownShelves(b *strings.Builder, shelves []shelf) {
	for _, s := range shelves {
		if len(s.Items) == 0 && !s.Pinned {
			continue
		}
		b.WriteString("\n" + md.H2(s.Title) + "\n")
		if s.Note != "" {
			b.WriteString(md.EscapeText(s.Note) + "\n\n")
		}
		if len(s.Items) == 0 {
			b.WriteString(md.EscapeText(s.Empty) + "\n")
		}
		writeItems(b, s.Items)
		if s.More.Key != "" {
			b.WriteString("\n" + md.Link("See all", storeURL+":"+s.More.Key) + "\n")
		}
	}
}

// renderPaged renders one page of a list under its title, chosen by the
// "page" query parameter (gnoweb's pager uses the same form), clamped to the
// valid range. base is the page's own path.
func renderPaged(l list, base, query string) string {
	values, _ := url.ParseQuery(query)
	n, _ := strconv.Atoi(values.Get("page"))
	if n < 1 {
		n = 1
	}
	members, pages := l.page(n)
	if n > pages {
		n = pages
		members, _ = l.page(n)
	}

	var b strings.Builder
	b.WriteString(md.H1(l.Title) + "\n")
	if len(members) == 0 {
		b.WriteString("No apps yet.\n")
	}
	writeItems(&b, members)
	if pages > 1 {
		pageURL := base + "?page="
		b.WriteString("\nPage " + strconv.Itoa(n) + " of " + strconv.Itoa(pages))
		if n > 1 {
			b.WriteString(" · " + md.Link("Previous", pageURL+strconv.Itoa(n-1)))
		}
		if n < pages {
			b.WriteString(" · " + md.Link("Next", pageURL+strconv.Itoa(n+1)))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n" + md.Link("Back to Explore", storeURL) + "\n")
	return b.String()
}

// writeItems lists listings, each linked to its own realm or package page.
func writeItems(b *strings.Builder, ls []*listing) {
	for _, l := range ls {
		href := strings.TrimPrefix(l.Path, runtime.ChainDomain())
		b.WriteString(md.BulletItem(md.Bold(md.Link(l.Title, href)) + " — " + md.EscapeText(l.Tagline)))
	}
}

// maxBlock is the most apps a Block lists.
const maxBlock = 8

// Block returns a markdown block another realm can embed in its Render: the
// first n apps (1 to 8) of the store list "latest" (as the store front's
// New), "top" (as its Top) or "trending", one link per line, then a link to
// the whole list. It returns "" for any other list. It reads state only and
// needs no cross call.
func Block(list string, n int) string {
	if n < 1 {
		n = 1
	} else if n > maxBlock {
		n = maxBlock
	}
	var items []*listing
	switch list {
	case "latest":
		items = newApps(n)
	case "top":
		items = topApps(n)
	case "trending":
		items = trending(runtime.ChainHeight(), n).items
	default:
		return ""
	}
	var b strings.Builder
	writeItems(&b, items)
	b.WriteString("\n" + md.Link("See all on Explore", storeURL+":"+list) + "\n")
	return b.String()
}

// Badge returns a markdown line for the shown app listed at path, with its
// ranked stars as the store counts them and a link to star it, or "". An
// app shows it in its own Render instead of a count of its own.
func Badge(path string) string {
	l := lookup(&byPath, path)
	if l == nil || !l.shownApp() {
		return ""
	}
	return "★ " + strconv.Itoa(l.rankedCount()) + " on Explore · " +
		md.Link("Star", storeURL+"$help&func=Star&slug="+l.Slug)
}
