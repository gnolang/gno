package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareUsesHTMLTemplate(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		want   string
	}{
		{"normal remote", "localhost:9999", "const ws = new WebSocket('ws://localhost:9999');"},
		{"xss'd remote", `localhost:9999');alert('pwned`, "const ws = new WebSocket('ws://localhost:9999&#39;);alert(&#39;pwned');"},
	}

	// As the code revolves, add more search patterns here.
	reWebsocket := regexp.MustCompile("const ws = new WebSocket[^\n]+")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			script, err := emitter.GenerateReloadScript(tt.remote)
			require.NoError(t, err, "GenerateReloadScript should not error")

			mdw := NewInjectorMiddleware([][]byte{script}, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Content-Type", "text/html")
				fmt.Fprintf(rw, "<body></body>")
			}))
			rec.Header().Set("Content-Type", "text/html")
			req := httptest.NewRequest("GET", "https://gno.land/example", nil)
			mdw.ServeHTTP(rec, req)

			targets := reWebsocket.FindAllString(rec.Body.String(), -1)
			require.True(t, len(targets) > 0)
			body := targets[0]
			require.Equal(t, body, tt.want)
		})
	}
}
