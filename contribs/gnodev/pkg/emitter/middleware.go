package emitter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
)

//go:embed static/hotreload.js
var reloadscript string

type middleware struct {
	remote   string
	muRemote sync.RWMutex

	next     http.Handler
	tmpl     *template.Template
	onceExec sync.Once
	script   []byte
}

// NewMiddleware creates an HTTP handler that acts as middleware. Its primary
// purpose is to intercept HTTP responses and inject a WebSocket client script
// into the body of HTML pages. This injection allows for dynamic content
// updates on the client side without requiring a page refresh.
func NewMiddleware(remote string, next http.Handler) http.Handler {
	tmpl := template.Must(template.New("reloadscript").
		Funcs(tmplFuncs).
		Parse(reloadscript))

	return &middleware{
		tmpl:     tmpl,
		remote:   remote,
		next:     next,
		onceExec: sync.Once{},
	}
}

type middlewareResponseWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func (m *middlewareResponseWriter) Write(b []byte) (int, error) {
	return m.buffer.Write(b)
}

type data struct {
	Remote       string
	ReloadEvents []events.Type
}

func (m *middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.muRemote.RLock()
	defer m.muRemote.RUnlock()

	// Create a buffer to hold the modified response
	buffer := bytes.NewBuffer(nil)

	// Create a ResponseWriter that writes to our buffer
	mw := &middlewareResponseWriter{
		ResponseWriter: rw,
		buffer:         buffer,
	}

	// Call the next handler, which writes to our buffer
	m.next.ServeHTTP(mw, req)

	// Check for any "text/html" answer
	content := mw.ResponseWriter.Header().Get("Content-Type")
	if !strings.Contains(content, "text/html") {
		rw.Write(buffer.Bytes())
		return
	}

	m.onceExec.Do(func() {
		script := &bytes.Buffer{}
		script.WriteString(`<script type="text/javascript">`)
		err := m.tmpl.Execute(script, &data{
			Remote: m.remote,
			ReloadEvents: []events.Type{
				events.EvtReload, events.EvtReset, events.EvtTxResult,
			},
		})
		if err != nil {
			panic("unable to execute template: " + err.Error())
		}
		script.WriteString("</script>")
		script.WriteString("</body>")
		m.script = script.Bytes()
	})

	// Inject the script before </body>
	updated := bytes.Replace(
		buffer.Bytes(),
		[]byte("</body>"),
		m.script, 1)

	rw.Write(updated)
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
