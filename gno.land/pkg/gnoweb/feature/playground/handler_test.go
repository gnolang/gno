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
	// pkgPath and expr are what the last Eval was asked.
	pkgPath, expr string
}

func (s *stubClient) Eval(_ context.Context, pkgPath, expr string) ([]byte, error) {
	s.pkgPath, s.expr = pkgPath, expr
	return s.evalResult, s.evalErr
}

// stubPackages is a PackageFetcher answering files or err, and recording
// the path it was asked for.
type stubPackages struct {
	files []File
	err   error
	asked string
}

func (s *stubPackages) PackageFiles(_ context.Context, pkgPath string) ([]File, error) {
	s.asked = pkgPath
	return s.files, s.err
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
			name:       "the chain domain is stripped",
			body:       `{"pkg_path":"gno.land/r/mock/path","expression":"Render(\"\")"}`,
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
			handler := New(Deps{Packages: &stubPackages{}, Client: client, Logger: discardLogger()}).EvalHandler()

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
			if tc.wantResult != "" {
				// The node is asked the request's package, without its
				// domain, and its expression.
				assert.Equal(t, "r/mock/path", client.pkgPath)
				assert.Equal(t, `Render("")`, client.expr)
			}
			assert.Contains(t, resp.Error, tc.wantError)
			if tc.wantError == "" {
				assert.Empty(t, resp.Error)
			}
			// Only the sentinel's message reaches the client.
			assert.NotContains(t, resp.Error, "10.0.0.1")
		})
	}
}

// TestGetPlaygroundViewCode covers the "code" shared-snippet paths. Each
// code holds a sentinel the default code does not, so a fallback to the
// default never passes for a decoded snippet.
func TestGetPlaygroundViewCode(t *testing.T) {
	t.Parallel()

	const sentinel = "package sentinel_42\n"
	require.NotContains(t, defaultCode, "sentinel_42")
	z := deflateBase64(t, []byte(sentinel))

	cases := []struct {
		name  string
		query url.Values
		want  string
	}{
		{"no code", url.Values{}, defaultCode},
		{"plain text", url.Values{"code": {sentinel}}, sentinel},
		{"plain text that parses as base64", url.Values{"code": {"test"}}, "test"},
		{"base64 without z stays text", url.Values{"code": {"cGFja2FnZSBtYWluCg=="}}, "cGFja2FnZSBtYWluCg=="},
		{"compressed", url.Values{"code": {z}, "z": {""}}, sentinel},
		{"compressed, cut short", url.Values{"code": {z[:len(z)-1]}, "z": {""}}, defaultCode},
		{"z on base64 that is no deflate", url.Values{"code": {"dGVzdA=="}, "z": {""}}, defaultCode},
		{"z on text that is no base64", url.Values{"code": {"package main"}, "z": {""}}, defaultCode},
		// The cap is written out, not read from the constant, so raising it
		// fails here: 64 KiB is what a link Share writes can hold.
		{"compressed up to 64 KiB", url.Values{"code": {deflateBase64(t, bytes.Repeat([]byte("a"), 64<<10))}, "z": {""}}, strings.Repeat("a", 64<<10)},
		// A few bytes of a link must not expand into a page of megabytes.
		{"compressed past 64 KiB", url.Values{"code": {deflateBase64(t, bytes.Repeat([]byte("a"), 64<<10+1))}, "z": {""}}, defaultCode},
		{"a link-sized megabyte", url.Values{"code": {deflateBase64(t, bytes.Repeat([]byte("a"), 1<<20))}, "z": {""}}, defaultCode},
		{"longer than Share writes", url.Values{"code": {strings.Repeat("a", maxSharedCodeLength+1)}}, defaultCode},
		{"compressed, longer than Share writes", url.Values{"code": {z + strings.Repeat("A", maxSharedCodeLength)}, "z": {""}}, defaultCode},
	}

	h := New(validDeps())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: tc.query})
			assert.Equal(t, http.StatusOK, status)
			assert.Equal(t, tc.want, extractPlaygroundViewData(t, v).InitialCode)
		})
	}
}

func TestCopyView(t *testing.T) {
	t.Parallel()

	t.Run("files are handed over one by one, sources only, in order", func(t *testing.T) {
		t.Parallel()

		// a.gno holds a line shaped like the share format's separator: a
		// copy never goes through that format, so it stays in a.gno.
		a := "package main\n\n// --- b.gno ---\nfunc A() {}\n"
		pkgs := &stubPackages{files: []File{
			{Name: "a.gno", Content: a},
			{Name: "README.md", Content: "# readme"},
			{Name: "b.gno", Content: "package main\n"},
			{Name: "gnomod.toml", Content: "module = \"main\"\n"},
		}}
		deps := validDeps()
		deps.Packages = pkgs
		h := New(deps)

		status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: url.Values{"from": {"gno.land/r/demo/foo"}, "file": {"b.gno"}}})
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "/r/demo/foo", pkgs.asked)

		data := extractPlaygroundViewData(t, v)
		assert.Equal(t, []File{
			{Name: "a.gno", Content: a},
			{Name: "b.gno", Content: "package main\n"},
			{Name: "gnomod.toml", Content: "module = \"main\"\n"},
		}, data.Files)
		assert.Empty(t, data.InitialCode)
		assert.Equal(t, "b.gno", data.DefaultFile)
		assert.Equal(t, "r/demo/foo", data.From)
		assert.Equal(t, deps.Domain, data.Domain)
		assert.Equal(t, "README.md", pkgs.files[1].Name, "the fetcher's slice is left as it was")

		// The page carries them as JSON, which reads back to the same files.
		var buf bytes.Buffer
		require.NoError(t, v.Render(&buf))
		page := buf.String()
		const open = `<script type="application/json" data-playground-target="files">`
		i := strings.Index(page, open)
		require.NotEqual(t, -1, i, "copy files script")
		j := strings.Index(page[i:], "</script>")
		var got []File
		require.NoError(t, json.Unmarshal([]byte(page[i+len(open):i+j]), &got))
		assert.Equal(t, data.Files, got)
		assert.NotContains(t, page, `data-playground-target="initial-code"`)
		// Run previews the realm it was copied from, by its own path.
		assert.Contains(t, page, `data-playground-from-value="r/demo/foo"`)
	})

	t.Run("a closing script tag in a file stays inside the JSON", func(t *testing.T) {
		t.Parallel()

		deps := validDeps()
		deps.Packages = &stubPackages{files: []File{{Name: "a.gno", Content: "// </script><script>alert(1)</script>"}}}
		_, v := New(deps).copyView(context.Background(), "/r/demo/foo", "")
		var buf bytes.Buffer
		require.NoError(t, v.Render(&buf))
		assert.NotContains(t, buf.String(), "<script>alert(1)")
	})

	errCases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", fmt.Errorf("%w: gone", ErrPackageNotFound), http.StatusNotFound},
		{"too large", fmt.Errorf("%w: 101 files", ErrPackageTooLarge), http.StatusRequestEntityTooLarge},
		{"timeout", fmt.Errorf("%w: deadline", ErrPackageTimeout), http.StatusRequestTimeout},
		{"anything else", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := validDeps()
			deps.Packages = &stubPackages{err: tc.err}
			status, v := New(deps).copyView(context.Background(), "/r/demo/foo", "")
			assert.Equal(t, tc.want, status)
			_, ok := v.Component.(*playgroundComponent)
			assert.False(t, ok, "an error is no playground")
		})
	}

	// Only a bare package path is copied: anything else is never fetched,
	// so a file's lines are not taken for file names.
	for _, from := range []string{"r/demo/foo/foo.gno", "r/demo/foo/README.md", "r/demo/foo:args", "r/demo/foo$source", "u/alice", "r/", "evil.com/x"} {
		t.Run("from "+from, func(t *testing.T) {
			t.Parallel()

			pkgs := &stubPackages{files: []File{{Name: "a.gno", Content: "package a"}}}
			deps := validDeps()
			deps.Packages = pkgs
			status, _ := New(deps).GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: url.Values{"from": {from}}})
			assert.Equal(t, http.StatusBadRequest, status)
			assert.Empty(t, pkgs.asked, "not fetched")
		})
	}
}

func TestParseFrom(t *testing.T) {
	t.Parallel()

	for from, want := range map[string]string{
		"r/demo/foo":          "/r/demo/foo",
		"/p/demo/foo":         "/p/demo/foo",
		"gno.land/r/demo/foo": "/r/demo/foo",
		"r/demo/foo/foo.gno":  "",
		"r/demo/foo:args":     "",
		"r/demo/foo$source":   "",
		"r/demo/foo?x=1":      "",
		"u/alice":             "",
		"r/":                  "",
		"r/demo/foo/":         "",
		"///evil.com/x":       "",
	} {
		u := ParseFrom("gno.land", from)
		if want == "" {
			assert.Nil(t, u, from)
			continue
		}
		require.NotNil(t, u, from)
		assert.Equal(t, want, u.Path, from)
	}
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
		Client:   &stubClient{evalResult: []byte("ok")},
		Packages: &stubPackages{},
		Logger:   discardLogger(),
		Limiter:  &denyLimiter{n: 2},
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
