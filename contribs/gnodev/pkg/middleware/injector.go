package middleware

import (
	"bytes"
	_ "embed"
	"net/http"
	"strings"
)

type injectorMiddleware struct {
	next     http.Handler
	toInject []byte
}

// NewInjectorMiddleware creates a new middleware that injects custom JavaScript
// scripts into HTML responses before the closing </body> tag.
func NewInjectorMiddleware(scripts [][]byte, next http.Handler) http.Handler {
	var concat bytes.Buffer

	// Concat all scripts into a single byte slice.
	for _, script := range scripts {
		concat.WriteString("<script type=\"text/javascript\">\n")
		concat.Write(script)
		concat.WriteString("</script>\n")
	}

	return &injectorMiddleware{
		next:     next,
		toInject: concat.Bytes(),
	}
}

type middlewareResponseWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func (m *middlewareResponseWriter) Write(b []byte) (int, error) {
	return m.buffer.Write(b)
}

func (m *injectorMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

	// Inject the scripts before </body>
	updated := bytes.Replace(
		buffer.Bytes(),
		[]byte("</body>"),
		m.toInject, 1)

	rw.Write(updated)
}
