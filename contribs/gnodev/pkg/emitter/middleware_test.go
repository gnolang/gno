package emitter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareUsesHTMLTemplate(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"normal path", "/_events", "const ws = new WebSocket((window.location.protocol === 'https:' ? 'wss://' : 'ws://') + window.location.host + '/_events');"},
		{"xss'd path", `/_events');alert('pwned`, "const ws = new WebSocket((window.location.protocol === 'https:' ? 'wss://' : 'ws://') + window.location.host + '/_events&#39;);alert(&#39;pwned');"},
	}

	// As the code revolves, add more search patterns here.
	reWebsocket := regexp.MustCompile("const ws = new WebSocket[^\n]+")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mdw := NewMiddleware(tt.path, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
