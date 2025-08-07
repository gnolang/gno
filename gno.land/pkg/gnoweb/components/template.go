package components

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/docparser"
)

//go:embed ui/*.html views/*.html layouts/*.html
var html embed.FS

var funcMap = template.FuncMap{}

var tmpl = template.New("web")

// rendererInstance is a global instance of the renderer
var rendererInstance interface{}

// SetRenderer sets the global renderer instance for syntax highlighting
func SetRenderer(renderer interface{}) {
	rendererInstance = renderer
}

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
	// highlightCode highlights code with a specific language
	funcs["highlightCode"] = func(code, language string) template.HTML {
		if rendererInstance == nil {
			return template.HTML(fmt.Sprintf(`<pre class="chroma">%s</pre>`, code))
		}
		
		// Use the HighlightCode method
		if r, ok := rendererInstance.(interface{ HighlightCode(string, string) (string, error) }); ok {
			if highlighted, err := r.HighlightCode(code, language); err == nil {
				return template.HTML(highlighted)
			} else {
				// Log error and fallback to plain text
				return template.HTML(fmt.Sprintf(`<pre class="chroma"><!-- Error: %s -->%s</pre>`, err.Error(), code))
			}
		}
		
		return template.HTML(fmt.Sprintf(`<pre class="chroma">%s</pre>`, code))
	}
	
	// dict creates a map from key-value pairs
	funcs["dict"] = func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict requires even number of arguments")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	}
	
	// printf formats a string like fmt.Sprintf
	funcs["printf"] = fmt.Sprintf
	funcs["formatDoc"] = formatDoc
}

// formatDoc formats documentation with code blocks using Chroma syntax highlighting
func formatDoc(doc string) template.HTML {
	if len(doc) > 100000 {
		return template.HTML("<div class=\"text-red-500\">Documentation too large (max 100KB)</div>")
	}
	
	blocks, err := docparser.ParseDocumentation(doc)
	if err != nil {
		return template.HTML(fmt.Sprintf("<div class=\"text-red-500\">Error parsing documentation: %s</div>", err.Error()))
	}
	
	if len(blocks) == 0 {
		return ""
	}
	
	// Execute template
	var buf bytes.Buffer
	tmpl.ExecuteTemplate(&buf, "ui/doc_content", map[string]interface{}{"Blocks": blocks})
	return template.HTML(buf.String())
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
