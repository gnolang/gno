package components

import (
	"bytes"
	"go/token"
	"html/template"
	"io"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// buildSymbols splits jdoc.Funcs into top-level functions and methods
// attached to their receiver TypeEntry. Unexported entries are dropped.
// Methods whose receiver isn't declared in jdoc.Types are silently skipped.
func buildSymbols(jdoc *doc.JSONDocumentation, render DocRenderer, pkgPath string) ([]FuncEntry, []TypeEntry) {
	if jdoc == nil {
		return nil, nil
	}

	typeTable := make(map[string]*TypeEntry, len(jdoc.Types))
	for _, t := range jdoc.Types {
		if !token.IsExported(t.Name) {
			continue
		}
		typeTable[t.Name] = &TypeEntry{
			Name:               t.Name,
			SignatureComponent: renderSignature(t.Type, render),
			Kind:               t.Kind,
			Doc:                renderDocString(t.Doc, render),
			AnchorID:           "type-" + t.Name,
			SourceURL:          buildSourceURL(pkgPath, t.File, t.Line),
		}
	}

	var topFuncs []FuncEntry
	for _, fn := range jdoc.Funcs {
		if !token.IsExported(fn.Name) {
			continue
		}
		entry := FuncEntry{
			Name:               fn.Name,
			SignatureComponent: renderSignature(fn.Signature, render),
			Doc:                renderDocString(fn.Doc, render),
			Crossing:           fn.Crossing,
			Receiver:           fn.Type,
			IsMethod:           fn.Type != "",
			AnchorID:           symbolAnchor(fn),
			SourceURL:          buildSourceURL(pkgPath, fn.File, fn.Line),
		}
		// Only realms (/r/) expose callable actions ($help); pure packages don't.
		if fn.Type == "" && fn.Name != "Render" && strings.HasPrefix(pkgPath, "/r/") {
			entry.ActionURL = pkgPath + "$help&func=" + fn.Name
		}
		if fn.Type == "" {
			topFuncs = append(topFuncs, entry)
			continue
		}
		if t, ok := typeTable[fn.Type]; ok {
			t.Methods = append(t.Methods, entry)
		}
	}

	out := make([]TypeEntry, 0, len(jdoc.Types))
	for _, t := range jdoc.Types {
		if entry := typeTable[t.Name]; entry != nil {
			out = append(out, *entry)
		}
	}
	return topFuncs, out
}

// isExportedValueDecl reports whether a value declaration group contains at
// least one exported name.
func isExportedValueDecl(v *doc.JSONValueDecl) bool {
	for _, vv := range v.Values {
		if token.IsExported(vv.Name) {
			return true
		}
	}
	return false
}

// buildSourceURL returns the deep link to a symbol's declaration site in the
// source view. Returns "" if file/line is not known — callers must check
// before emitting links.
func buildSourceURL(pkgPath, file string, line int) string {
	if file == "" || line <= 0 {
		return ""
	}
	return pkgPath + "$source&file=" + file + "#L" + strconv.Itoa(line)
}

// symbolAnchor returns a stable anchor id for a function or method.
func symbolAnchor(fn *doc.JSONFunc) string {
	if fn.Type != "" {
		return "method-" + fn.Type + "-" + fn.Name
	}
	return "func-" + fn.Name
}

// renderSignature syntax-highlights a Gno signature snippet as HTML.
// Returns nil for empty signatures; falls back to HTML-escaped text on renderer failure.
func renderSignature(sig string, render DocRenderer) Component {
	if strings.TrimSpace(sig) == "" {
		return nil
	}
	var buf bytes.Buffer
	if err := render.RenderSource(&buf, "sig.gno", []byte(sig)); err != nil {
		return rawHTMLComponent(template.HTMLEscapeString(sig))
	}
	return rawHTMLComponent(buf.String())
}

// renderDocString renders a doc string through the DocRenderer into a Component.
// A nil or empty doc returns nil so the template can use {{ with .Doc }}.
// Renderer failures degrade to a plain HTML-escaped string.
func renderDocString(src string, render DocRenderer) Component {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	var buf bytes.Buffer
	if err := render.RenderDocumentation(&buf, []byte(src)); err != nil {
		escaped := template.HTMLEscapeString(src)
		return rawHTMLComponent(escaped)
	}
	return rawHTMLComponent(buf.String())
}

// rawHTMLComponent wraps pre-rendered HTML (safe by construction) as a Component.
type rawHTMLComponent string

func (s rawHTMLComponent) Render(w io.Writer) error {
	_, err := io.WriteString(w, string(s))
	return err
}

// buildValues flattens jdoc.Values preserving source order and groups names.
// Unexported declarations are dropped (jdoc may include unexported symbols).
func buildValues(jdoc *doc.JSONDocumentation, render DocRenderer, pkgPath string) []ValueGroup {
	if jdoc == nil || len(jdoc.Values) == 0 {
		return nil
	}
	out := make([]ValueGroup, 0, len(jdoc.Values))
	for i, v := range jdoc.Values {
		if !isExportedValueDecl(v) {
			continue
		}
		kind := "var"
		if v.Const {
			kind = "const"
		}
		names := make([]string, 0, len(v.Values))
		for _, vv := range v.Values {
			names = append(names, vv.Name)
		}
		anchorID := "value-"
		if len(v.Values) > 0 {
			anchorID += v.Values[0].Name
		} else {
			anchorID += kind + "-" + strconv.Itoa(i)
		}
		out = append(out, ValueGroup{
			Kind:               kind,
			Names:              strings.Join(names, ", "),
			SignatureComponent: renderSignature(v.Signature, render),
			Doc:                renderDocString(v.Doc, render),
			AnchorID:           anchorID,
			SourceURL:          buildSourceURL(pkgPath, v.File, v.Line),
		})
	}
	return out
}
