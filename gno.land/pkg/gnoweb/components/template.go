package components

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

//go:embed ui/*.html views/*.html layouts/*.html
var html embed.FS

var funcMap = template.FuncMap{}

var tmpl = template.New("web")

func registerCommonFuncs(funcs template.FuncMap) {
	// NOTE: this method does NOT escape HTML, use with caution
	// Render Component element into raw html element
	funcs["render"] = func(comp Component) (template.HTML, error) {
		var buf bytes.Buffer
		if err := comp.Render(&buf); err != nil {
			return "", fmt.Errorf("unable to render component: %w", err)
		}

		return template.HTML(buf.String()), nil //nolint:gosec
	}
	funcs["queryHas"] = func(vals url.Values, key string) bool {
		if vals == nil {
			return false
		}

		return vals.Has(key)
	}
	funcs["FormatRelativeTime"] = FormatRelativeTimeSince
	// add returns the sum of two integers — used by recursive templates that
	// track depth (e.g. the state explorer tree).
	funcs["add"] = func(a, b int) int { return a + b }
	funcs["sub"] = func(a, b int) int { return a - b }
	// derefInt dereferences a `*int` for template arithmetic. Go's
	// html/template `with` does NOT auto-deref pointers, so comparing
	// `*int` against `int` directly raises "invalid type for
	// comparison" at execute time. Used by the state-explorer "+N more"
	// CTA where StateNode.Length is `*int` (nil = unknown count).
	funcs["derefInt"] = func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	}
	// truncOID surfaces a long ObjectID/Hash as `head…tail` (preserving
	// the `:N` suffix when present) so chips and sidebar rows stay
	// scannable without losing the full value (kept in `title`/copy).
	funcs["truncOID"] = TruncOID
	// headingForKind picks the appropriate column-header label for a
	// children grid based on the parent node's Kind. Keeps the binding
	// label consistent between nested levels (struct→Field, map→Key,
	// slice/array→Index, default→Field).
	funcs["headingForKind"] = func(kind string) string {
		switch kind {
		case "map":
			return "Key"
		case "slice", "array":
			return "Index"
		default:
			return "Field"
		}
	}
	// kindIconID picks the SVG sprite ID (without the `ico-` prefix)
	// for a node's Kind+Type. Used as a leading visual hint next to
	// the field name so users recognise the shape at a glance —
	// rendered as `<svg><use href="#ico-kind-..."/></svg>` so styling,
	// theming, and accessibility align with the rest of gnoweb's
	// icon system. Symbols are defined in `ui/icons.html`.
	funcs["kindIconID"] = func(kind, t string) string {
		switch kind {
		case "primitive":
			switch {
			case strings.Contains(t, "string"):
				return "kind-string"
			case strings.Contains(t, "bool"):
				return "kind-bool"
			default:
				return "kind-number"
			}
		case "struct":
			return "kind-struct"
		case "map":
			return "kind-map"
		case "slice", "array":
			return "kind-slice"
		case "pointer":
			return "kind-pointer"
		case "func":
			return "kind-func"
		case "closure":
			return "kind-closure"
		case "ref":
			return "kind-ref"
		case "nil":
			return "kind-nil"
		case "package":
			return "kind-package"
		case "type":
			return "kind-type"
		case "interface":
			return "kind-interface"
		default:
			return "kind-unknown"
		}
	}
	// oidShort: trailing `:N` when id and ref share the same 40-char
	// hashlet, full id otherwise. Avoids rendering near-identical
	// Owner/OID pairs in the audit chips.
	funcs["oidShort"] = ShortenOID
	// dict creates a map from key-value pairs for passing multiple values to templates
	funcs["dict"] = func(kv ...any) (map[string]any, error) {
		if len(kv)%2 != 0 {
			return nil, fmt.Errorf("dict requires an even number of arguments")
		}
		result := make(map[string]any)
		for i := 0; i < len(kv); i += 2 {
			key, ok := kv[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			result[key] = kv[i+1]
		}
		return result, nil
	}
}

func init() {
	// Register templates functions
	registerCommonFuncs(funcMap)
	registerHelpFuncs(funcMap)
	tmpl.Funcs(funcMap)

	// Parse templates
	var err error
	tmpl, err = tmpl.ParseFS(html, "layouts/*.html", "ui/*.html", "views/*.html")
	if err != nil {
		panic("unable to parse embed tempalates: " + err.Error())
	}
}
