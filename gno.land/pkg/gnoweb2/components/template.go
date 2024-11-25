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
	"noescape_string": func(in string) template.HTML {
		return template.HTML(in)
	},
	"noescape_bytes": func(in []byte) template.HTML {
		return template.HTML(in)
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



// func (c *ComponentsTemplate) ExecuteTemplate(ctx context.Context, wr io.Writer, name string, data any) {
// 	c.Template.Funcs(RenderFunc(ctx)).ExecuteTemplate(wr, name, data)
// }

// func Templates() (tmpl *template.Template) {
// 	return tmpl
// }

// type Slot struct {
// 	ID   string
// 	Html string
// }

// func StreamShadowRoot(w http.ResponseWriter, ss <-chan Slot) error {
// 	w.(http.Flusher).Flush()
// 	for slot := range ss {
// 		if err := tmpl.ExecuteTemplate(w, "slot", nil); err != nil {
// 			return fmt.Errorf("unable to execute slot template: %w", err)
// 		}
// 	}
// }

// func templateMain() {
// 	tmpl, err := template.ParseFS(gohtml, "*.gohtml")
// 	if err != nil {
// 		panic(err)
// 	}

// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		data := struct {
// 			Meta PageMetadata
// 			Body template.HTML
// 		}{
// 			Meta: PageMetadata{
// 				Title:       "My Page Title",
// 				Description: "A short description of my page.",
// 				Canonical:   "http://example.com",
// 				Image:       "http://example.com/image.jpg",
// 				URL:         "http://example.com",
// 			},
// 			Body: template.HTML("<p>This is the main content of the page.</p>"),
// 		}

// 		err := tmpl.ExecuteTemplate(w, "index", data)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 		}
// 	})

// 	http.ListenAndServe(":8080", nil)
// }
