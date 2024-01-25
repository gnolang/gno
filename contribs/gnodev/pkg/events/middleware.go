package events

import (
	"bytes"
	_ "embed"
	"net/http"
	"strings"
	"sync"
	"text/template"
)

//go:embed static/hotreload.js
var reloadscript string

type Middleware struct {
	remote   string
	muRemote sync.RWMutex

	next     http.Handler
	tmpl     *template.Template
	onceExec *sync.Once
	script   []byte
}

func NewMiddleware(remote string, next http.Handler) *Middleware {
	tmpl := template.Must(template.New("reloadscript").
		Funcs(tmplFuncs).
		Parse(reloadscript))

	return &Middleware{
		tmpl:     tmpl,
		remote:   remote,
		next:     next,
		onceExec: &sync.Once{},
	}
}

type middlewareResponseWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func (m *middlewareResponseWriter) Write(b []byte) (int, error) {
	return m.buffer.Write(b)
}

func (m *Middleware) UpdateRemote(remote string) {
	m.muRemote.Lock()
	m.remote = remote
	m.onceExec = &sync.Once{}
	m.muRemote.Unlock()
}

type data struct {
	Remote       string
	ReloadEvents []EventType
}

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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
			Remote:       m.remote,
			ReloadEvents: []EventType{EvtReload, EvtReset, EvtTxResult},
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
