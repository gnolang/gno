package playground

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
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
	filesErr   error
	fileBodies map[string][]byte
}

func (s *stubClient) ListFiles(context.Context, string) ([]string, error) {
	return s.files, s.filesErr
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

// deflate compresses bytes with DEFLATE (RFC 1951) and returns the base64 of it.
func deflateBase64(t *testing.T, b []byte) string {
	t.Helper()

	var buf bytes.Buffer
	zw, err := flate.NewWriter(&buf, flate.BestCompression)
	require.NoError(t, err)
	_, err = zw.Write(b)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// extractPlaygroundViewData extracts data from a playground view.
func extractPlaygroundViewData(t *testing.T, v *components.View) PlaygroundData {
	t.Helper()

	c, ok := v.Component.(*playgroundComponent)
	require.True(t, ok, "unexpected component type %T", v.Component)
	data, ok := c.data.(PlaygroundData)
	require.True(t, ok, "unexpected data type %T", c.data)
	return data
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

// TestGetPlaygroundViewCode covers the "code" shared-snippet paths.
func TestGetPlaygroundViewCode(t *testing.T) {
	t.Parallel()

	h := New(validDeps())
	indexData := &components.IndexData{}

	t.Run("default when no code", func(t *testing.T) {
		t.Parallel()

		status, v := h.GetPlaygroundView(&weburl.GnoURL{Query: url.Values{}}, indexData)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, defaultCode, extractPlaygroundViewData(t, v).InitialCode)
	})

	t.Run("plain base64", func(t *testing.T) {
		t.Parallel()

		code := "package main // hello"
		q := url.Values{"code": {base64.StdEncoding.EncodeToString([]byte(code))}}
		status, v := h.GetPlaygroundView(&weburl.GnoURL{Query: q}, indexData)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, code, extractPlaygroundViewData(t, v).InitialCode)
	})

	t.Run("compressed round-trip", func(t *testing.T) {
		t.Parallel()

		code := "package main\n\nfunc Render(path string) string { return \"hi\" }\n"
		q := url.Values{"code": {deflateBase64(t, []byte(code))}, "z": {""}}
		status, v := h.GetPlaygroundView(&weburl.GnoURL{Query: q}, indexData)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, code, extractPlaygroundViewData(t, v).InitialCode)
	})

	t.Run("decompression bomb is rejected", func(t *testing.T) {
		t.Parallel()

		// ~8 MiB of zeros compresses to a few KB; well over the 1 MiB ceiling.
		bomb := deflateBase64(t, bytes.Repeat([]byte{0}, 8<<20))
		q := url.Values{"code": {bomb}, "z": {""}}
		status, v := h.GetPlaygroundView(&weburl.GnoURL{Query: q}, indexData)
		assert.Equal(t, http.StatusOK, status)

		// The over-limit payload must not be adopted as InitialCode,
		// the guard sets the default code instead.
		got := extractPlaygroundViewData(t, v).InitialCode
		assert.Equal(t, got, defaultCode, "view must use default code")
	})
}

func TestGetForkView(t *testing.T) {
	t.Parallel()

	t.Run("multiple files concatenated with headers", func(t *testing.T) {
		t.Parallel()

		deps := validDeps()
		deps.Client = &stubClient{
			files: []string{"a.gno", "gnomod.toml"},
			fileBodies: map[string][]byte{
				"a.gno":       []byte("package main\n"),
				"gnomod.toml": []byte("module = \"main\"\n"),
			},
		}
		h := New(deps)

		status, v := h.GetForkView(context.Background(), &weburl.GnoURL{Path: "/r/demo/foo"})
		require.Equal(t, http.StatusOK, status)

		code := extractPlaygroundViewData(t, v).InitialCode

		// First header has no leading newline; the second one does
		assert.True(t, strings.HasPrefix(code, "// --- a.gno ---\n\n"), "first header should not have a leading newline")
		assert.Contains(t, code, "\n// --- gnomod.toml ---\n\n")
		assert.Contains(t, code, "package main\n")
		assert.Contains(t, code, "module = \"main\"\n")

		// a.gno appears before gnomod.toml (list order preserved)
		assert.Less(t, strings.Index(code, "a.gno"), strings.Index(code, "gnomod.toml"))
	})

	t.Run("non-source files are filtered out", func(t *testing.T) {
		t.Parallel()

		deps := validDeps()
		deps.Client = &stubClient{
			files: []string{"a.gno", "image.png", "README.md", "gnomod.toml"},
			fileBodies: map[string][]byte{
				"a.gno":       []byte("package main\n"),
				"image.png":   []byte("binary"),
				"README.md":   []byte("# readme"),
				"gnomod.toml": []byte("module = \"main\"\n"),
			},
		}
		h := New(deps)

		status, v := h.GetForkView(context.Background(), &weburl.GnoURL{Path: "/r/demo/foo"})
		require.Equal(t, http.StatusOK, status)

		code := extractPlaygroundViewData(t, v).InitialCode
		assert.Contains(t, code, "// --- a.gno ---")
		assert.Contains(t, code, "// --- gnomod.toml ---")
		assert.NotContains(t, code, "image.png")
		assert.NotContains(t, code, "README.md")
	})

	t.Run("oversized source is rejected", func(t *testing.T) {
		t.Parallel()

		// Two files whose combined size exceeds maxForkCodeSize.
		big := bytes.Repeat([]byte("x"), maxForkCodeSize)
		deps := validDeps()
		deps.Client = &stubClient{
			files: []string{"a.gno", "b.gno"},
			fileBodies: map[string][]byte{
				"a.gno": big,
				"b.gno": []byte("package main\n"),
			},
		}
		h := New(deps)

		status, v := h.GetForkView(context.Background(), &weburl.GnoURL{Path: "/r/demo/foo"})
		assert.Equal(t, http.StatusRequestEntityTooLarge, status)

		// The error path returns a status error component, not a playground view
		_, ok := v.Component.(*playgroundComponent)
		assert.False(t, ok, "oversized path must not return a playground component")
	})

	t.Run("fail listing package files", func(t *testing.T) {
		t.Parallel()

		deps := validDeps()
		deps.Client = &stubClient{filesErr: errors.New("boom")}
		h := New(deps)

		status, v := h.GetForkView(context.Background(), &weburl.GnoURL{Path: "/r/demo/foo"})
		assert.Equal(t, http.StatusBadRequest, status)

		// The error path returns a status error component, not a playground view
		_, ok := v.Component.(*playgroundComponent)
		assert.False(t, ok, "error path must not return a playground component")
	})

	t.Run("playground data fields are populated", func(t *testing.T) {
		t.Parallel()

		deps := validDeps()
		deps.Client = &stubClient{
			files:      []string{"a.gno"},
			fileBodies: map[string][]byte{"a.gno": []byte("package main\n")},
		}
		h := New(deps)

		pkgPath := "/r/demo/foo"
		q := url.Values{"file": {"a.gno"}}
		status, v := h.GetForkView(context.Background(), &weburl.GnoURL{Path: pkgPath, Query: q})
		require.Equal(t, http.StatusOK, status)

		data := extractPlaygroundViewData(t, v)
		assert.Equal(t, path.Join(deps.Domain, pkgPath), data.ForkFrom)
		assert.Equal(t, "a.gno", data.DefaultFile)
		assert.Equal(t, deps.Remote, data.Remote)
		assert.Equal(t, deps.ChainId, data.ChainId)
		assert.Equal(t, deps.Domain, data.Domain)
	})
}

// TestDecodeCompressedCode unit-tests the bounded DEFLATE decoder directly.
func TestDecodeCompressedCode(t *testing.T) {
	t.Parallel()

	t.Run("valid small payload", func(t *testing.T) {
		t.Parallel()

		want := "some gno source"
		var buf bytes.Buffer
		zw, _ := flate.NewWriter(&buf, flate.BestCompression)
		_, _ = zw.Write([]byte(want))
		require.NoError(t, zw.Close())

		got, ok := decodeCompressedCode(buf.Bytes())
		require.True(t, ok)
		assert.Equal(t, want, got)
	})

	t.Run("over-limit payload rejected", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		zw, _ := flate.NewWriter(&buf, flate.BestCompression)
		_, _ = zw.Write(bytes.Repeat([]byte{0}, maxDecompressedCodeSize+1))
		require.NoError(t, zw.Close())

		_, ok := decodeCompressedCode(buf.Bytes())
		assert.False(t, ok, "payload exceeding the ceiling must be rejected")
	})

	t.Run("invalid deflate data rejected", func(t *testing.T) {
		t.Parallel()

		_, ok := decodeCompressedCode([]byte("not deflate data"))
		assert.False(t, ok)
	})
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
