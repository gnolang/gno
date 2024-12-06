package components

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io"
	"net/url"
)

//go:embed *.gohtml
var gohtml embed.FS

var funcMap = template.FuncMap{
	// NOTE: this method does NOT escape HTML, use with caution
	"noescape_string": func(in string) template.HTML {
		return template.HTML(in) //nolint:gosec
	},
	// NOTE: this method does NOT escape HTML, use with caution
	"noescape_bytes": func(in []byte) template.HTML {
		return template.HTML(in) //nolint:gosec
	},
	"queryHas": func(vals url.Values, key string) bool {
		if vals == nil {
			return false
		}

		return vals.Has(key)
	},
}

var tmpl = template.New("web").Funcs(funcMap)

func init() {
	registerHelpFuncs(funcMap)
	tmpl.Funcs(funcMap)

	var err error
	tmpl, err = tmpl.ParseFS(gohtml, "*.gohtml")
	if err != nil {
		panic("unable to parse embed tempalates: " + err.Error())
	}
}

type Component func(ctx context.Context, tmpl *template.Template, w io.Writer) error

func (c Component) Render(ctx context.Context, w io.Writer) error {
	return RenderComponent(ctx, w, c)
}

func RenderComponent(ctx context.Context, w io.Writer, c Component) error {
	var render *template.Template
	funcmap := template.FuncMap{
		"render": func(cf Component) (string, error) {
			var buf bytes.Buffer
			if err := cf(ctx, render, &buf); err != nil {
				return "", err
			}

			return buf.String(), nil
		},
	}

	render = tmpl.Funcs(funcmap)
	return c(ctx, render, w)
}

type StatusData struct {
	Message string
}

func RenderStatusComponent(w io.Writer, message string) error {
	return tmpl.ExecuteTemplate(w, "status", StatusData{
		Message: message,
	})
}
