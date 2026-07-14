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
	funcs["hasPrefix"] = strings.HasPrefix
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
	// truncMiddle shortens long opaque strings (e.g. bech32 addresses) for the
	// sidebar: keeps `keep` runes on each side joined by an ellipsis.
	funcs["truncMiddle"] = func(s string, keep int) string {
		r := []rune(s)
		if keep <= 0 || len(r) <= keep*2+1 {
			return s
		}
		return string(r[:keep]) + "…" + string(r[len(r)-keep:])
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
		panic("unable to parse embed templates: " + err.Error())
	}
}
