package playground

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubClient is a minimal ClientAdapter for handler tests — every method
// returns canned data so the API path can be exercised without importing
// the gnoweb package (which would create a test-time cycle).
type stubClient struct {
	evalResult []byte
	evalErr    error
	docResult  *doc.JSONDocumentation
	docErr     error
	files      []string
	fileBodies map[string][]byte
}

func (s *stubClient) ListFiles(context.Context, string) ([]string, error) {
	return s.files, nil
}

func (s *stubClient) File(_ context.Context, _, filename string) ([]byte, error) {
	body, ok := s.fileBodies[filename]
	if !ok {
		return nil, errors.New("file not found")
	}
	return body, nil
}

func (s *stubClient) Doc(context.Context, string) (*doc.JSONDocumentation, error) {
	return s.docResult, s.docErr
}

func (s *stubClient) Eval(context.Context, string) ([]byte, error) {
	return s.evalResult, s.evalErr
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestHandlerPlaygroundEval tests the POST /_/api/eval handler directly.
func TestHandlerPlaygroundEval(t *testing.T) {
	t.Parallel()

	h := New(Deps{
		Client:  &stubClient{evalResult: []byte("mock eval")},
		Logger:  discardLogger(),
		Domain:  "gno.land",
		Remote:  "http://localhost:26657",
		ChainId: "test",
	})

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

	handler := h.EvalHandler()
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

	deps := validDeps()
	deps.Client = &stubClient{
		docResult: &doc.JSONDocumentation{
			Funcs: []*doc.JSONFunc{
				{Name: "Hello", Signature: "Hello() string"},
				{Name: "method", Type: "MyType", Signature: "method()"},
			},
		},
	}
	h := New(deps)

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

	handler := h.FuncsHandler()
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

// TestRateLimiter tests that the per-IP rate limiter enforces burst
// limits. Uses a custom (burst=2, refill=10s) bucket so the test does
// not race the production +1/3s refill rate.
func TestRateLimiter(t *testing.T) {
	t.Parallel()

	h := &Handler{
		deps: Deps{
			Client: &stubClient{evalResult: []byte("ok")},
			Logger: discardLogger(),
			Domain: "gno.land",
		},
		limiter: newRateLimiter(2, 10*time.Second),
	}
	handler := h.EvalHandler()

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
