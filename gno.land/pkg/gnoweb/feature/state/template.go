package state

import (
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

//go:embed templates/*.html
var templateFS embed.FS

// funcMap mirrors components.funcMap (package-private there) so the
// shared helper subset renders identically to legacy state.html.
var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },

	// derefInt: html/template's `with` does not auto-deref *int, so
	// arithmetic on .Length needs this helper.
	"derefInt": func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	},

	"truncOID": TruncOID,
	"oidShort": ShortenOID,

	"headingForKind": func(kind string) string {
		switch kind {
		case KindMap:
			return "Key"
		case KindSlice, KindArray:
			return "Index"
		default:
			return "Field"
		}
	},

	"kindGroup": func(kind string) string {
		switch kind {
		case KindStruct, KindMap, KindSlice, KindArray, KindPointer, KindRef:
			return "state"
		case KindFunc, KindClosure:
			return "code"
		case KindType, KindInterface:
			return "types"
		default:
			return "other"
		}
	},

	"kindIconID": func(kind, t string) string {
		switch kind {
		case KindPrimitive:
			switch {
			case strings.Contains(t, "string"):
				return "kind-string"
			case strings.Contains(t, "bool"):
				return "kind-bool"
			default:
				return "kind-number"
			}
		case KindStruct:
			return "kind-struct"
		case KindMap:
			return "kind-map"
		case KindSlice, KindArray:
			return "kind-slice"
		case KindPointer:
			return "kind-pointer"
		case KindFunc:
			return "kind-func"
		case KindClosure:
			return "kind-closure"
		case KindRef:
			return "kind-ref"
		case KindNil:
			return "kind-nil"
		case KindPackage:
			return "kind-package"
		case KindType:
			return "kind-type"
		case KindInterface:
			return "kind-interface"
		default:
			return "kind-unknown"
		}
	},

	// sourceHref carries ?height=N so time-travel context survives the
	// hop from a state card to the source tab.
	"sourceHref": func(pkgPath, file string, line int, height int64) template.URL {
		wq := url.Values{"source": {""}, "file": {file}}
		if height > 0 {
			wq.Set("height", strconv.FormatInt(height, 10))
		}
		u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
		href := u.EncodeWebURL()
		if line > 0 {
			href += "#L" + strconv.Itoa(line)
		}
		return template.URL(href) //nolint:gosec
	},

	// hx-get / permalink builders — see helpers.go.
	"stateFragNodeHref":   stateFragNodeHref,
	"stateFragSourceHref": stateFragSourceHref,
	"stateObjectHref":     stateObjectHref,
	"stateSourceHref":     stateSourceHref,

	"dict": func(kv ...any) (map[string]any, error) {
		if len(kv)%2 != 0 {
			return nil, fmt.Errorf("dict requires an even number of arguments")
		}
		result := make(map[string]any, len(kv)/2)
		for i := 0; i < len(kv); i += 2 {
			key, ok := kv[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			result[key] = kv[i+1]
		}
		return result, nil
	},
}

// Pre-parsed templates for the state feature. mustParse panics at
// init on a malformed template — misconfiguration surfaces immediately,
// not on the first request.
//
// _nodes.html (state/nodes, state/node, state/source-details) is parsed
// into BOTH the page and the node-fragment template sets so an
// htmx-loaded fragment renders with the exact same recursive markup +
// .row/--depth CSS as the server-rendered tree.
var (
	PageTemplate       = mustParse("renderPage", "templates/page.html", "templates/_nodes.html")
	FragNodeTemplate   = mustParse("fragNode", "templates/frag_node.html", "templates/_nodes.html")
	FragSourceTemplate = mustParse("fragSource", "templates/frag_source.html")
	FragErrorTemplate  = mustParse("fragError", "templates/frag_error.html")
)

func mustParse(name string, paths ...string) *template.Template {
	t, err := template.New(name).Funcs(funcMap).ParseFS(templateFS, paths...)
	if err != nil {
		panic("state: parse " + paths[0] + ": " + err.Error())
	}
	return t
}
