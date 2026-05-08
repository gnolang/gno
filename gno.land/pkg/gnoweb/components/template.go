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
	funcs["truncOID"] = truncOID
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
	// kindGlyph picks a small Unicode glyph for a node's Kind+Type.
	// Used as a leading visual hint next to the field name so users
	// recognise the shape at a glance — `T` strings, `#` numbers,
	// `⊞` structs, `≡` maps, etc. Pure CSS-renderable, no SVG asset.
	funcs["kindGlyph"] = func(kind, t string) string {
		switch kind {
		case "primitive":
			switch {
			case strings.Contains(t, "string"):
				return "T"
			case strings.Contains(t, "bool"):
				return "◐"
			default:
				return "#"
			}
		case "struct":
			return "⊞"
		case "map":
			return "≡"
		case "slice", "array":
			return "[ ]"
		case "pointer":
			return "→"
		case "func":
			return "ƒ"
		case "closure":
			return "λ"
		case "ref":
			return "◇"
		case "nil":
			return "∅"
		case "package":
			return "⌥"
		case "type":
			return "T:"
		case "interface":
			return "?"
		default:
			return "·"
		}
	}
	// oidShort returns the trailing `:N` of an ObjectID when it shares
	// its 40-char hashlet with `ref`. Otherwise the full id. Used to
	// avoid rendering near-identical Owner/OID pairs in the audit chips.
	funcs["oidShort"] = func(id, ref string) string {
		i, j := strings.IndexByte(id, ':'), strings.IndexByte(ref, ':')
		if i > 0 && j > 0 && id[:i] == ref[:j] {
			return id[i:]
		}
		return id
	}
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
