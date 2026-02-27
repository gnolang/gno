package emitter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
)

//go:embed static/hotreload.js
var reloadscript string

type data struct {
	Remote       string
	ReloadEvents []events.Type
}

// GenerateReloadScript generates a JavaScript script that can be injected
// into HTML pages using a middleware to enable hot reloading functionality
// using WebSockets.
func GenerateReloadScript(remote string) ([]byte, error) {
	tmpl := template.Must(template.New("reloadscript").
		Funcs(tmplFuncs).
		Parse(reloadscript),
	)

	script := &bytes.Buffer{}
	if err := tmpl.Execute(script, &data{
		Remote: remote,
		ReloadEvents: []events.Type{
			events.EvtReload, events.EvtReset, events.EvtTxResult,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to execute template: %w", err)
	}

	return script.Bytes(), nil
}

var tmplFuncs = template.FuncMap{
	"json": func(obj any) string {
		raw, err := json.Marshal(obj)
		if err != nil {
			panic(fmt.Errorf("marshal error: %w", err))
		}

		return string(raw)
	},
}
