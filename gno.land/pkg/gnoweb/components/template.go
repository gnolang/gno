package components

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
)

//go:embed ui/*.html views/*.html layouts/*.html
var html embed.FS

var funcMap = template.FuncMap{}

var tmpl = template.New("web")

func registerCommonFuncs(funcs template.FuncMap) {
	// NOTE: this method does NOT escape HTML, use with caution
	funcs["noescape_string"] = func(in string) template.HTML {
		return template.HTML(in) //nolint:gosec
	}
	// NOTE: this method does NOT escape HTML, use with caution
	funcs["noescape_bytes"] = func(in []byte) template.HTML {
		return template.HTML(in) //nolint:gosec
	}
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
	funcs["truncateMiddle"] = func(s string, visibleChars int) string {
		if len(s) <= visibleChars*2+3 {
			return s
		}
		if visibleChars < 1 {
			return s[:3] + "..."
		}
		return s[:visibleChars] + "..." + s[len(s)-visibleChars:]
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
