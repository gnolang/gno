package gnoweb_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postEval sends one /_/api/eval request from remoteAddr with the given
// extra headers and returns the status.
func postEval(t *testing.T, h http.Handler, remoteAddr string, headers map[string]string) int {
	t.Helper()

	body := `{"pkg_path":"r/demo/counter","expression":"Render(\"\")"}`
	req := httptest.NewRequest(http.MethodPost, "/_/api/eval", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.RemoteAddr = remoteAddr
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr.Code
}

// TestPlaygroundEvalLimiter checks the eval budget and that the client is
// identified the way every other gnoweb limiter identifies it.
func TestPlaygroundEvalLimiter(t *testing.T) {
	t.Parallel()

	newEval := func(t *testing.T, trusted ...string) http.Handler {
		t.Helper()
		cfg := newTestHandlerConfig(t, gnoweb.NewMockClient())
		cfg.StateRateLimitTrustedProxies = trusted
		h, err := gnoweb.NewHTTPHandler(slog.New(slog.DiscardHandler), cfg)
		require.NoError(t, err)
		return h.Playground.EvalHandler()
	}

	countLimited := func(n int, send func(i int) int) (limited int) {
		for i := range n {
			if send(i) == http.StatusTooManyRequests {
				limited++
			}
		}
		return limited
	}

	t.Run("a load and a reload of a dozen getters fit", func(t *testing.T) {
		t.Parallel()
		eval := newEval(t)
		limited := countLimited(24, func(int) int {
			return postEval(t, eval, "192.0.2.1:1234", nil)
		})
		assert.Zero(t, limited)
	})

	t.Run("an untrusted peer is one client", func(t *testing.T) {
		t.Parallel()
		eval := newEval(t)
		limited := countLimited(100, func(i int) int {
			return postEval(t, eval, "192.0.2.1:1234", map[string]string{
				"X-Real-IP": fmt.Sprintf("10.0.1.%d", i),
			})
		})
		assert.NotZero(t, limited)
	})

	t.Run("readers behind a trusted proxy each get their own bucket", func(t *testing.T) {
		t.Parallel()
		eval := newEval(t, "127.0.0.1/32")
		limited := countLimited(100, func(i int) int {
			return postEval(t, eval, "127.0.0.1:1234", map[string]string{
				"X-Real-IP": fmt.Sprintf("10.0.1.%d", i),
			})
		})
		assert.Zero(t, limited)
	})
}
