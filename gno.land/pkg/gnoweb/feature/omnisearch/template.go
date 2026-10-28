package omnisearch

import (
	"embed"
	"html/template"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

//go:embed templates/*.html
var templateFS embed.FS

// funcMap holds only what these templates need. It deliberately does not
// import the components package's map: that one is package-private there, and
// a shared map would couple the feature's rendering to the core's.
var funcMap = template.FuncMap{
	"isIndexer": func(s Source) bool { return s == SourceIndexer },
	"errText": func(err error) string {
		if err == nil {
			return ""
		}
		return err.Error()
	},
	// groupAt hands a group its position, which anchors the rail's links.
	"groupAt": func(i int, g Group) groupRow { return groupRow{Index: i, Group: g} },
	// keywordLists is the keywords popover's content.
	"keywordLists": keywordLists,
	// shortDetail cuts an opaque detail (an address, a hash) in its middle,
	// where neither end a reader recognizes it by is lost. Prose (a synopsis,
	// a render snippet) is left whole; the stylesheet ellipsizes it.
	"shortDetail": func(s string) string {
		if strings.ContainsRune(s, ' ') {
			return s
		}
		return components.TruncMiddle(s, 6)
	},
	// shortQuery cuts each long value of a query in its middle, so an address
	// typed after `account:` fits beside the title. The page keeps the whole
	// query in the element's title.
	"shortQuery": func(q string) string {
		words := strings.Fields(q)
		for i, w := range words {
			key, val, ok := strings.Cut(w, ":")
			if !ok {
				key, val = "", w
			} else {
				key += ":"
			}
			words[i] = key + components.TruncMiddle(val, 8)
		}
		return strings.Join(words, " ")
	},
	// rowIcon picks a row's icon from where it leads: the kind of page is
	// what a reader scans a mixed listing for.
	"rowIcon": func(href template.URL) string {
		switch h := string(href); {
		case h == "":
			return "ico-tx-link"
		case strings.HasPrefix(h, "/u/"):
			return "ico-user-link"
		case strings.HasPrefix(h, "/r/"):
			return "ico-realm"
		case strings.HasPrefix(h, "/p/"):
			return "ico-pure"
		}
		return "ico-file"
	},
}

// groupRow is the payload of omnisearch/group.
type groupRow struct {
	Index int
	Group Group
}

// hintRow is the payload of omnisearch/hint.
type hintRow struct {
	Hint, Help string
	Indexer    bool
	Href       template.URL
}

// hintList is one heading of the keywords popover and its rows. Provenance
// is a row's, as in the omnibar: each indexer row carries its own pill.
type hintList struct {
	Label string
	Rows  []hintRow
}

// angles sets a hint's placeholder as the omnibar does: `in:‹path›`.
var angles = strings.NewReplacer("<", "‹", ">", "›")

// keywordLists groups the keywords as the omnibar does: those that work
// anywhere under keywordGroups, then those about a package. On a package
// page, a keyword that takes no argument links to its own answer there.
func keywordLists(d SearchData) []hintList {
	selRow := func(s *Selector, pkgPath string) hintRow {
		row := hintRow{Hint: angles.Replace(s.Hint), Help: s.Help, Indexer: s.Source == SourceIndexer}
		if s.Bare {
			row.Href = keywordHref(pkgPath, s.Name)
		}
		return row
	}
	var lists []hintList
	add := func(label string, rows []hintRow) {
		if len(rows) == 0 {
			return
		}
		lists = append(lists, hintList{Label: label, Rows: rows})
	}
	for _, g := range keywordGroups {
		var rows []hintRow
		for _, s := range d.Selectors {
			if s.Scope != ScopePackage && s.Group == g.Key {
				rows = append(rows, selRow(s, ""))
			}
		}
		for _, f := range d.Filters {
			if f.Group == g.Key {
				rows = append(rows, hintRow{Hint: angles.Replace(f.Hint), Help: f.Help})
			}
		}
		add(g.Label, rows)
	}
	var pkg []hintRow
	for _, s := range d.Selectors {
		if s.Scope == ScopePackage {
			pkg = append(pkg, selRow(s, d.PkgPath))
		}
	}
	add("On a package page", pkg)
	return lists
}

// pageTemplate is parsed once at init: a malformed template surfaces
// immediately rather than on the first request. The rail's mobile toggle is
// the shared partial, so it opens as every other rail does.
var pageTemplate = template.Must(template.Must(
	template.New("renderSearch").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"),
).ParseFS(components.SharedPartialsFS(), "expend_label.html"))
