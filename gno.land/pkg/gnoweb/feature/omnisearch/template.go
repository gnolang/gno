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
	// hintRow is one keyword of the rail. On a package page, a keyword that
	// takes no argument links to its own answer there.
	"hintRow": func(sel *Selector, pkgPath string) hintRow {
		row := hintRow{Hint: sel.Hint, Label: sel.Label, Indexer: sel.Source == SourceIndexer}
		if sel.Bare {
			row.Href = keywordHref(pkgPath, sel.Name)
		}
		return row
	},
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
	Hint, Label string
	Indexer     bool
	Href        template.URL
}

// pageTemplate is parsed once at init: a malformed template surfaces
// immediately rather than on the first request. The rail's mobile toggle is
// the shared partial, so it opens as every other rail does.
var pageTemplate = template.Must(template.Must(
	template.New("renderSearch").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"),
).ParseFS(components.SharedPartialsFS(), "expend_label.html"))
