package gnoweb

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandlerPlaygroundEval tests the POST /_/api/eval handler directly.
func TestHandlerPlaygroundEval(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io_discard{}, nil))
	cli := NewMockClient(&MockPackage{
		Domain: "gno.land",
		Path:   "/r/mock/path",
		Files:  map[string]string{"mock.gno": `package mock`},
	})
	handler := handlerPlaygroundEval(logger, cli, "gno.land", "http://localhost:26657")

	cases := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantResult string
		wantError  string
	}{
		{
			name:       "valid eval",
			method:     http.MethodPost,
			body:       `{"pkg_path":"r/mock/path","expression":"Render(\"\")"}`,
			wantStatus: http.StatusOK,
			wantResult: "mock eval",
		},
		{
			name:       "missing pkg_path",
			method:     http.MethodPost,
			body:       `{"expression":"Render(\"\")"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "required",
		},
		{
			name:       "missing expression",
			method:     http.MethodPost,
			body:       `{"pkg_path":"r/mock/path"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "required",
		},
		{
			name:       "invalid json",
			method:     http.MethodPost,
			body:       `not json`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid",
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, "/_/api/eval", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantStatus == http.StatusOK {
				var resp map[string]string
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				if tc.wantResult != "" {
					assert.Contains(t, resp["result"], tc.wantResult)
				}
				if tc.wantError != "" {
					assert.Contains(t, resp["error"], tc.wantError)
				}
			} else if tc.wantError != "" {
				assert.Contains(t, rr.Body.String(), tc.wantError)
			}
		})
	}
}

// TestHandlerPlaygroundFuncs tests the GET /_/api/funcs handler directly.
func TestHandlerPlaygroundFuncs(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io_discard{}, nil))
	cli := NewMockClient(&MockPackage{
		Domain: "gno.land",
		Path:   "/r/mock/path",
		Files:  map[string]string{"mock.gno": `package mock`},
		Functions: []*doc.JSONFunc{
			{Name: "Hello", Signature: "Hello() string"},
			{Name: "method", Type: "MyType", Signature: "method()"},
		},
	})
	handler := handlerPlaygroundFuncs(logger, cli)

	cases := []struct {
		name       string
		path       string
		method     string
		wantStatus int
		check      func(t *testing.T, body string)
	}{
		{
			name:       "valid path",
			path:       "/_/api/funcs?path=/r/mock/path",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			check: func(t *testing.T, body string) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal([]byte(body), &resp))
				funcs, ok := resp["functions"].([]any)
				assert.True(t, ok)
				found := false
				for _, f := range funcs {
					fm := f.(map[string]any)
					if fm["name"] == "Hello" {
						found = true
					}
					assert.NotEqual(t, "method", fm["name"], "methods should be filtered")
				}
				assert.True(t, found, "Hello function should be present")
			},
		},
		{
			name:       "missing path param",
			path:       "/_/api/funcs",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			path:       "/_/api/funcs?path=/r/mock/path",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.check != nil {
				tc.check(t, rr.Body.String())
			}
		})
	}
}

// TestRateLimiter tests that the per-IP rate limiter enforces burst limits.
func TestRateLimiter(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io_discard{}, nil))
	cli := NewMockClient(&MockPackage{
		Domain: "gno.land",
		Path:   "/r/mock/path",
		Files:  map[string]string{"mock.gno": `package mock`},
	})

	// Burst of 2, refill every 10 seconds (won't refill during test).
	rl := newRateLimiter(2, 10*time.Second)
	h := &playgroundAPIHandler{
		logger:  logger,
		client:  cli,
		domain:  "gno.land",
		limiter: rl,
	}
	handler := http.HandlerFunc(h.serveEval)

	body := `{"pkg_path":"r/mock/path","expression":"Render(\"\")"}`
	ip := "192.0.2.1:1234"

	// First two requests should succeed (burst=2).
	for i := range 2 {
		req := httptest.NewRequest(http.MethodPost, "/_/api/eval", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should succeed", i+1)
	}

	// Third request should be rate-limited.
	req := httptest.NewRequest(http.MethodPost, "/_/api/eval", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code, "third request should be rate-limited")
	assert.Contains(t, rr.Body.String(), "rate limit")
}

// io_discard is an io.Writer that discards all output, used for test loggers.
type io_discard struct{}

func (io_discard) Write(p []byte) (int, error) { return len(p), nil }
