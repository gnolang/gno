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
	"shortDetail":  shortDetail,
	"rowIcon":      rowIcon,
}

// detailEnds is how much of each end an opaque detail keeps. With the
// "deployed by " before it, the cut fits .b-omni-meta .detail's 26ch, which
// would otherwise take the end this cut keeps.
const detailEnds = 6

// shortDetail cuts the address or hash a detail ends in, in its middle, where
// neither end a reader recognizes it by is lost. Prose (a synopsis, a render
// snippet) is left whole; the stylesheet ellipsizes it.
func shortDetail(r Result) string {
	if !r.Opaque {
		return r.Detail
	}
	i := strings.LastIndexByte(r.Detail, ' ') + 1
	return r.Detail[:i] + components.TruncMiddle(r.Detail[i:], detailEnds)
}

// rowIcon picks a row's icon: what it names when its link does not say, else
// where it leads, the kind of page being what a reader scans a listing for.
func rowIcon(r Result) string {
	switch r.Kind {
	case KindTx:
		return "ico-tx-link"
	case KindBlock:
		return "ico-chain"
	}
	switch h := string(r.Href); {
	case strings.HasPrefix(h, "/u/"):
		return "ico-user-link"
	case strings.HasPrefix(h, "/r/"):
		return "ico-realm"
	case strings.HasPrefix(h, "/p/"), r.Kind == KindImport:
		return "ico-pure"
	}
	return "ico-file"
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
// anywhere under keywordGroups, and those about a package, the ones that take
// no argument leading, first on its page, where they link to their own answer.
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
		for _, f := range filterKeywords {
			if f.Group == g.Key {
				rows = append(rows, hintRow{Hint: angles.Replace(f.Hint), Help: f.Help})
			}
		}
		add(g.Label, rows)
	}
	var pkg []hintRow
	for _, bare := range []bool{true, false} {
		for _, s := range d.Selectors {
			if s.Scope == ScopePackage && s.Bare == bare {
				pkg = append(pkg, selRow(s, d.PagePath))
			}
		}
	}
	if d.PagePath == "" {
		add("On a package page", pkg)
		return lists
	}
	return append([]hintList{{Label: "On this package", Rows: pkg}}, lists...)
}

// pageTemplate is parsed once at init: a malformed template surfaces
// immediately rather than on the first request. The rail's mobile toggle is
// the shared partial, so it opens as every other rail does.
var pageTemplate = template.Must(template.Must(
	template.New("renderSearch").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"),
).ParseFS(components.SharedPartialsFS(), "expend_label.html"))
