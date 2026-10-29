package playground

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubClient is a minimal ClientAdapter for handler tests — every method
// returns canned data so the API path can be exercised without importing
// the gnoweb package (which would create a test-time cycle).
type stubClient struct {
	evalResult []byte
	evalErr    error
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

func (s *stubClient) Eval(context.Context, string, string) ([]byte, error) {
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

	const valid = `{"pkg_path":"r/mock/path","expression":"Render(\"\")"}`
	cases := []struct {
		name        string
		method      string
		contentType string // "" means application/json
		headers     map[string]string
		body        string
		client      *stubClient // nil means a client answering "mock eval"
		wantStatus  int
		wantResult  string
		wantError   string
		wantAllow   string
	}{
		{
			name:       "valid eval",
			body:       valid,
			wantStatus: http.StatusOK,
			wantResult: "mock eval",
		},
		{
			name:        "json with a charset",
			contentType: "application/json; charset=utf-8",
			body:        valid,
			wantStatus:  http.StatusOK,
			wantResult:  "mock eval",
		},
		{
			name:       "missing pkg_path",
			body:       `{"expression":"Render(\"\")"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "required",
		},
		{
			name:       "missing expression",
			body:       `{"pkg_path":"r/mock/path"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "required",
		},
		{
			name:       "invalid json",
			body:       `not json`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid",
		},
		{
			name:       "body over the cap",
			body:       `{"pkg_path":"r/mock/path","expression":"` + strings.Repeat("x", maxEvalBodyBytes) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantError:  "too large",
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method not allowed",
			wantAllow:  http.MethodPost,
		},
		{
			name:        "text/plain body",
			contentType: "text/plain",
			body:        valid,
			wantStatus:  http.StatusUnsupportedMediaType,
			wantError:   "application/json",
		},
		{
			name:       "cross-site fetch",
			headers:    map[string]string{"Sec-Fetch-Site": "cross-site"},
			body:       valid,
			wantStatus: http.StatusForbidden,
			wantError:  "cross-origin",
		},
		{
			name:       "foreign origin",
			headers:    map[string]string{"Origin": "https://evil.example"},
			body:       valid,
			wantStatus: http.StatusForbidden,
			wantError:  "cross-origin",
		},
		{
			name:       "same-origin fetch",
			headers:    map[string]string{"Sec-Fetch-Site": "same-origin", "Origin": "http://example.com"},
			body:       valid,
			wantStatus: http.StatusOK,
			wantResult: "mock eval",
		},
		{
			name:       "chain error is an answer",
			body:       valid,
			client:     &stubClient{evalErr: errors.New("name NoSuchThing not declared")},
			wantStatus: http.StatusOK,
			wantError:  "NoSuchThing not declared",
		},
		{
			name:       "node unreachable",
			body:       valid,
			client:     &stubClient{evalErr: fmt.Errorf("%w: dial tcp 10.0.0.1:26657", ErrEvalUnavailable)},
			wantStatus: http.StatusBadGateway,
			wantError:  ErrEvalUnavailable.Error(),
		},
		{
			name:       "node timeout",
			body:       valid,
			client:     &stubClient{evalErr: fmt.Errorf("%w: deadline", ErrEvalTimeout)},
			wantStatus: http.StatusGatewayTimeout,
			wantError:  ErrEvalTimeout.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := tc.client
			if client == nil {
				client = &stubClient{evalResult: []byte("mock eval")}
			}
			handler := New(Deps{Client: client, Logger: discardLogger()}).EvalHandler()

			method := tc.method
			if method == "" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, "http://example.com/_/api/eval", strings.NewReader(tc.body))
			contentType := tc.contentType
			if contentType == "" {
				contentType = "application/json"
			}
			req.Header.Set("Content-Type", contentType)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
			assert.Equal(t, tc.wantAllow, rr.Header().Get("Allow"))

			var resp evalResponse
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
			assert.Equal(t, tc.wantResult, resp.Result)
			assert.Contains(t, resp.Error, tc.wantError)
			if tc.wantError == "" {
				assert.Empty(t, resp.Error)
			}
			// Only the sentinel's message reaches the client.
			assert.NotContains(t, resp.Error, "10.0.0.1")
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

		status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: url.Values{}}, indexData)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, defaultCode, extractPlaygroundViewData(t, v).InitialCode)
	})

	t.Run("plain base64", func(t *testing.T) {
		t.Parallel()

		code := "package main // hello"
		q := url.Values{"code": {base64.StdEncoding.EncodeToString([]byte(code))}}
		status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: q}, indexData)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, code, extractPlaygroundViewData(t, v).InitialCode)
	})

	t.Run("compressed round-trip", func(t *testing.T) {
		t.Parallel()

		code := "package main\n\nfunc Render(path string) string { return \"hi\" }\n"
		q := url.Values{"code": {deflateBase64(t, []byte(code))}, "z": {""}}
		status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: q}, indexData)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, code, extractPlaygroundViewData(t, v).InitialCode)
	})

	t.Run("decompression bomb is rejected", func(t *testing.T) {
		t.Parallel()

		// ~8 MiB of zeros compresses to a few KB; well over the 1 MiB ceiling.
		bomb := deflateBase64(t, bytes.Repeat([]byte{0}, 8<<20))
		q := url.Values{"code": {bomb}, "z": {""}}
		status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: q}, indexData)
		assert.Equal(t, http.StatusOK, status)

		// The over-limit payload must not be adopted as InitialCode,
		// the guard sets the default code instead.
		got := extractPlaygroundViewData(t, v).InitialCode
		assert.Equal(t, got, defaultCode, "view must use default code")
	})
}

func TestCopyView(t *testing.T) {
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

		status, v := h.copyView(context.Background(), "/r/demo/foo", "")
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

		status, v := h.copyView(context.Background(), "/r/demo/foo", "")
		require.Equal(t, http.StatusOK, status)

		code := extractPlaygroundViewData(t, v).InitialCode
		assert.Contains(t, code, "// --- a.gno ---")
		assert.Contains(t, code, "// --- gnomod.toml ---")
		assert.NotContains(t, code, "image.png")
		assert.NotContains(t, code, "README.md")
	})

	t.Run("oversized source is rejected", func(t *testing.T) {
		t.Parallel()

		// Two files whose combined size exceeds maxCopyCodeSize.
		big := bytes.Repeat([]byte("x"), maxCopyCodeSize)
		deps := validDeps()
		deps.Client = &stubClient{
			files: []string{"a.gno", "b.gno"},
			fileBodies: map[string][]byte{
				"a.gno": big,
				"b.gno": []byte("package main\n"),
			},
		}
		h := New(deps)

		status, v := h.copyView(context.Background(), "/r/demo/foo", "")
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

		status, v := h.copyView(context.Background(), "/r/demo/foo", "")
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

		status, v := h.copyView(context.Background(), "/r/demo/foo", "a.gno")
		require.Equal(t, http.StatusOK, status)

		data := extractPlaygroundViewData(t, v)
		assert.Equal(t, "a.gno", data.DefaultFile)
		assert.Equal(t, deps.Domain, data.Domain)
	})
}

func TestRelPath(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"r/demo/foo":          "r/demo/foo",
		"/r/demo/foo":         "r/demo/foo",
		"gno.land/r/demo/foo": "r/demo/foo",
		"///evil.com/x":       "evil.com/x",
		"gno.land//r/demo":    "r/demo",
	} {
		assert.Equal(t, want, RelPath("gno.land", in), in)
	}
}

func TestCopyURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/_/play?from=r%2Fdemo%2Ffoo", CopyURL("/r/demo/foo", ""))
	assert.Equal(t, "/_/play?file=a.gno&from=r%2Fdemo%2Ffoo", CopyURL("/r/demo/foo", "a.gno"))
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

// denyLimiter refuses every request after the first n.
type denyLimiter struct{ n int }

func (l *denyLimiter) AllowRequest(*http.Request) bool {
	l.n--
	return l.n >= 0
}

// TestEvalLimiter checks that serveEval charges the configured limiter and
// answers 429 once it refuses, and that a request refused for its
// Content-Type or origin costs the reader nothing.
func TestEvalLimiter(t *testing.T) {
	t.Parallel()

	h := New(Deps{
		Client:  &stubClient{evalResult: []byte("ok")},
		Logger:  discardLogger(),
		Limiter: &denyLimiter{n: 2},
	})
	handler := h.EvalHandler()

	body := `{"pkg_path":"r/mock/path","expression":"Render(\"\")"}`
	requests := []map[string]string{
		{"Content-Type": "text/plain"},
		{"Content-Type": "application/json", "Sec-Fetch-Site": "cross-site"},
		{"Content-Type": "application/json"},
		{"Content-Type": "application/json"},
		{"Content-Type": "application/json"},
	}
	codes := make([]int, len(requests))
	for i, headers := range requests {
		req := httptest.NewRequest(http.MethodPost, "/_/api/eval", strings.NewReader(body))
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		codes[i] = rr.Code
	}
	assert.Equal(t, []int{
		http.StatusUnsupportedMediaType, http.StatusForbidden,
		http.StatusOK, http.StatusOK, http.StatusTooManyRequests,
	}, codes)
}
