package store

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// pageSize is the realm's page size for categories and lists (api/v1
// contract): with one's count it bounds which pages exist.
const pageSize = 24

// Meta is what a store view sets around itself: the page's status and its
// head title and description. Both texts come from fixed strings and the
// realm's closed category labels only, never from a listing: what is shared
// in a link preview is not a listing's to choose.
type Meta struct {
	Status      int
	Title       string
	Description string
}

// listName is gnoweb's own name for a list of the store front: the title
// that heads its page and its head meta, and its label in the switcher.
// Fixed strings: the realm's list titles name its shelves only.
type listName struct{ key, title, label string }

// lists are the lists the store front offers, in the order its switcher
// shows them.
var lists = []listName{
	{"latest", "New apps", "New"},
	{"top", "Top apps", "Top"},
	{"trending", "Trending this week", "Trending"},
	{"updated", "Recently updated", "Recently updated"},
}

func namedList(key string) (listName, bool) {
	i := slices.IndexFunc(lists, func(l listName) bool { return l.key == key })
	if i < 0 {
		return listName{}, false
	}
	return lists[i], true
}

func isList(key string) bool {
	_, ok := namedList(key)
	return ok
}

// The leads of the two lenses, on the page and in its head description.
const (
	introLine = "Apps you can read. Everything here runs on gno.land and its code is public: try an app, or read how it works first."
	buildLine = "Packages and services to build on, and the people who make them."
)

// View renders the store realm's Content tab for the render paths it owns:
// "" (the front page), "c/<key>" (a category), the lists the front page
// opens onto ("top", "latest", "trending", "updated") and "build". It returns
// nil for anything else, so gnoweb falls back to the realm's own markdown:
// other tabs, other render paths and markdown clients see the realm as it
// is. Under "c/", a key that is not a category the front page offers, valid
// or not, and a page of a category or list that does not exist are drawn as
// a store page that says so, with status 404, without querying the node.
// gnoweb calls it on the Content tab only, after every other tab dispatch.
// A store API failure also returns nil: it must never break the realm page,
// and is logged once when loaded (see decode), never shown, as it can carry
// the node's address.
func (h *Handler) View(ctx context.Context, u *weburl.GnoURL) (*components.View, Meta) {
	if u.Path != h.deps.RealmPath {
		return nil, Meta{}
	}
	switch args := u.Args; {
	case args == "":
		return h.serveHome(ctx)
	case args == "build":
		return h.serveBuild(ctx)
	case strings.HasPrefix(args, "c/"):
		return h.servePage(ctx, u, pagedCategory, args[len("c/"):])
	case isList(args):
		return h.servePage(ctx, u, pagedList, args)
	}
	return nil, Meta{}
}

func (h *Handler) meta(title, desc string) Meta {
	return Meta{Status: http.StatusOK, Title: title + " · Explore", Description: desc}
}

func (h *Handler) serveHome(ctx context.Context) (*components.View, Meta) {
	res, err := h.home(ctx)
	if err != nil {
		return nil, Meta{} // logged once by decode
	}
	return h.ok("store/home", h.buildHome(res)),
		Meta{Status: http.StatusOK, Title: "Explore · " + h.deps.Domain, Description: introLine}
}

func (h *Handler) serveBuild(ctx context.Context) (*components.View, Meta) {
	res, err := h.build(ctx)
	if err != nil {
		return nil, Meta{} // logged once by decode
	}
	return h.ok("store/build", h.buildBuild(res)), h.meta("Build", buildLine)
}

// servePage serves one page of a category or of a list (kind as in paged).
// It only queries the realm for one the cached front page offers, and a page
// that exists: arbitrary keys or pages never reach the node, nor fill the
// cache. A category page is headed by the front page's validated label and
// count, a list page by gnoweb's own name for it.
func (h *Handler) servePage(ctx context.Context, u *weburl.GnoURL, kind, key string) (*components.View, Meta) {
	home, err := h.home(ctx)
	if err != nil {
		return nil, Meta{} // logged once by decode
	}
	data := pageData{Nav: catNav{Items: home.Categories}}
	title, count, ok := offered(home, kind, key)
	page, valid := pageParam(u, count)
	if !ok || !valid {
		data.Nav.NoCurrent = true
		data.Title, data.Missing = "Not found", true
		data.Empty = "There is no such page in Explore. Pick a category above, or go back to all apps."
		return h.ok("store/page", data), Meta{Status: http.StatusNotFound, Title: "Not found · Explore", Description: introLine}
	}
	var m Meta
	if kind == pagedCategory {
		data.Nav.Active = key
		data.Title = title
		m = h.meta(title+" apps", title+" apps on "+h.deps.Domain+", each with its code public: try one, or read how it works first.")
	} else {
		name, _ := namedList(key)
		data.Nav.NoCurrent = true
		data.Lists = listTabs(home, key)
		data.Title = name.title
		m = h.meta(name.title, introLine)
	}
	res, err := h.paged(ctx, kind, key, page)
	if err != nil {
		return nil, Meta{} // an error is logged once by decode
	}
	data.Note = appCount(count)
	data.Cards = h.cards(ptrs(res.Listings), categoryLabels(home.Categories))
	if kind == pagedCategory {
		// The page is the category: its cards need not say it.
		for i := range data.Cards {
			data.Cards[i].Category = category{}
		}
		data.Empty = "No " + title + " apps yet."
	} else {
		data.Empty = "Nothing here yet."
	}
	data.Page, data.Pages = res.Page, res.Pages
	if page > 1 {
		data.Prev = page - 1
	}
	if page < res.Pages {
		data.Next = page + 1
	}
	return h.ok("store/page", data), m
}

// listTabs are the lists the front page offers, current one marked.
func listTabs(home *homeResponse, current string) []listTab {
	var out []listTab
	for _, l := range lists {
		if _, _, ok := offered(home, pagedList, l.key); ok {
			out = append(out, listTab{Key: l.key, Label: l.label, Current: l.key == current})
		}
	}
	return out
}

func (h *Handler) ok(tmpl string, data any) *components.View {
	return &components.View{
		Type:      ViewType,
		Component: pageComponent{tmpl: h.templates, name: tmpl, data: data},
	}
}

// pageParam reads ?page=N (1 when absent) and reports whether that page
// exists for a category or a list of count apps.
func pageParam(u *weburl.GnoURL, count int) (int, bool) {
	page := 1
	if q := u.Query.Get("page"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil {
			return 0, false
		}
		page = n
	}
	pages := max(1, (count+pageSize-1)/pageSize)
	return page, page >= 1 && page <= pages
}

// offered returns the title and count of a category or a list the front
// page offers.
func offered(home *homeResponse, kind, key string) (title string, count int, ok bool) {
	if kind == pagedCategory {
		for _, c := range home.Categories {
			if c.Key == key {
				return c.Label, c.Count, true
			}
		}
		return "", 0, false
	}
	for _, s := range home.Shelves {
		for _, l := range s.More {
			if l.Key == key {
				return l.Title, l.Count, true
			}
		}
	}
	return "", 0, false
}

func appCount(n int) string {
	if n == 1 {
		return "1 app"
	}
	return strconv.Itoa(n) + " apps"
}
