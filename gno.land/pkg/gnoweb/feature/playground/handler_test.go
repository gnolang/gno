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
	"golang.org/x/net/html"
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

// servedCode reads the code a page serves, as the controller does: each
// <pre> of the editor that names a file, or the shared snippet's <pre>, as
// a browser parses it.
func servedCode(t *testing.T, page string) (files []File, initial string) {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(page))
	require.NoError(t, err)
	attr := func(n *html.Node, key string) (string, bool) {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val, true
			}
		}
		return "", false
	}
	var text func(n *html.Node) string
	text = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		var b strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			b.WriteString(text(c))
		}
		return b.String()
	}
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode || n.Data != "pre" {
			continue
		}
		switch target, _ := attr(n, "data-playground-target"); target {
		case "file":
			name, _ := attr(n, "data-playground-name-value")
			files = append(files, File{Name: name, Content: text(n)})
		case "initial-code":
			initial = text(n)
		}
	}
	return files, initial
}

// extractPlaygroundViewData extracts data from a playground view.
func extractPlaygroundViewData(t *testing.T, v *components.View) PlaygroundData {
	t.Helper()

	c, ok := v.Component.(*components.FeatureComponent)
	require.True(t, ok, "unexpected component type %T", v.Component)
	data, ok := c.Data.(PlaygroundData)
	require.True(t, ok, "unexpected data type %T", c.Data)
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
		// One meaning per form: without &z nothing is decoded, not even
		// what Share would have written with it.
		{"compressed without z stays text", url.Values{"code": {z}}, z},
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
			// Served once, as text, which the controller reads back.
			var buf bytes.Buffer
			require.NoError(t, v.Render(&buf))
			_, initial := servedCode(t, buf.String())
			assert.Equal(t, tc.want, initial)
			assert.NotContains(t, buf.String(), "<textarea", "the code is in the page once")
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

		// The page serves each file once, as text under its name, which a
		// reader without JS reads and the controller reads back as is.
		var buf bytes.Buffer
		require.NoError(t, v.Render(&buf))
		page := buf.String()
		got, initial := servedCode(t, page)
		assert.Equal(t, data.Files, got)
		assert.Empty(t, initial)
		for _, f := range data.Files {
			assert.Contains(t, page, `<p class="b-playground-file u-font-mono">`+f.Name+"</p>")
		}
		assert.NotContains(t, page, "application/json", "the files are in the page once")
		// Run previews the realm it was copied from, by its own path.
		assert.Contains(t, page, `data-playground-from-value="r/demo/foo"`)
	})

	// Markup stays text, and so do a leading blank line, which HTML drops
	// right after <pre>, and the characters the text format would read.
	t.Run("a file reads back as it is", func(t *testing.T) {
		t.Parallel()

		files := []File{
			{Name: "a.gno", Content: "// </pre></script><script>alert(1)</script>\n"},
			{Name: "b.gno", Content: "\n\npackage b // & \"'<>\n"},
			{Name: "c.gno", Content: ""},
		}
		deps := validDeps()
		deps.Packages = &stubPackages{files: files}
		_, v := New(deps).copyView(context.Background(), "/r/demo/foo", "")
		var buf bytes.Buffer
		require.NoError(t, v.Render(&buf))
		assert.NotContains(t, buf.String(), "<script>alert(1)")
		got, _ := servedCode(t, buf.String())
		assert.Equal(t, files, got)
	})

	// Only the sources a copy shows count toward its cap, and its page
	// holds them once: a package of 1 MiB of sources is a page of about
	// 1 MiB, not two.
	t.Run("the cap counts the sources shown, which the page holds once", func(t *testing.T) {
		t.Parallel()

		half := strings.Repeat("x", maxCopySourceBytes/2)
		readme := File{Name: "README.md", Content: strings.Repeat("r", 4*maxCopySourceBytes)}
		cases := []struct {
			name  string
			files []File
			want  int
		}{
			{"at the cap, with a larger README", []File{{Name: "a.gno", Content: half}, {Name: "b.gno", Content: half}, readme}, http.StatusOK},
			{"one byte past the cap", []File{{Name: "a.gno", Content: half}, {Name: "b.gno", Content: half + "x"}}, http.StatusRequestEntityTooLarge},
		}
		for _, tc := range cases {
			deps := validDeps()
			deps.Packages = &stubPackages{files: tc.files}
			status, v := New(deps).copyView(context.Background(), "/r/demo/foo", "")
			require.Equal(t, tc.want, status, tc.name)
			if status != http.StatusOK {
				continue
			}
			var buf bytes.Buffer
			require.NoError(t, v.Render(&buf))
			assert.Less(t, buf.Len(), maxCopySourceBytes+16<<10, tc.name)
		}
		// Written out, not read from the constant: raising the cap fails here.
		assert.Equal(t, 1<<20, maxCopySourceBytes)
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
			assert.NotEqual(t, components.PlaygroundViewType, v.Type, "an error is no playground")
		})
	}

	// A copy that timed out and an eval that did say which one did.
	t.Run("timeouts are told apart", func(t *testing.T) {
		t.Parallel()
		assert.NotEqual(t, ErrPackageTimeout.Error(), ErrEvalTimeout.Error())
	})

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

// TestRunLabel checks the page names what Run does before any script runs:
// Preview on a copy of a realm that declares Render, Run locally otherwise.
func TestRunLabel(t *testing.T) {
	t.Parallel()

	const render = "package foo\n\nfunc Render(path string) string { return \"\" }\n"
	cases := []struct {
		name, pkg, file, code string
		preview               bool
	}{
		{"realm with Render", "/r/demo/foo", "foo.gno", render, true},
		{"unnamed path", "/r/demo/foo", "foo.gno", "package foo\n\nfunc Render(string) string { return \"\" }\n", true},
		{"realm without Render", "/r/demo/foo", "foo.gno", "package foo\n", false},
		{"Render in a comment", "/r/demo/foo", "foo.gno", "package foo\n// func Render(path string) string\n", false},
		{"pure package with Render", "/p/demo/foo", "foo.gno", render, false},
		// gnoweb renders a realm with Render(path string) string, from
		// the package's own files: a test's Render is not the realm's.
		{"Render in a test", "/r/demo/foo", "foo_test.gno", render, false},
		{"Render in a filetest", "/r/demo/foo", "z_filetest.gno", render, false},
		{"crossing Render", "/r/demo/foo", "foo.gno", "package foo\n\nfunc Render(cur realm, path string) string { return \"\" }\n", false},
		{"Render with no result", "/r/demo/foo", "foo.gno", "package foo\n\nfunc Render(path string) {}\n", false},
		{"Render of two params", "/r/demo/foo", "foo.gno", "package foo\n\nfunc Render(path string, n int) string { return \"\" }\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := validDeps()
			deps.Packages = &stubPackages{files: []File{{Name: tc.file, Content: tc.code}}}
			status, v := New(deps).copyView(context.Background(), tc.pkg, "")
			require.Equal(t, http.StatusOK, status)
			var buf bytes.Buffer
			require.NoError(t, v.Render(&buf))
			page := buf.String()
			// The controller reads the decision and the texts it switches
			// to on the first edit off the button: it holds none of its own.
			if tc.preview {
				assert.Contains(t, page, "<span>Preview</span>")
				assert.Contains(t, page, `title="Show what `+deps.Domain+tc.pkg+` renders (Ctrl/Cmd+Enter)"`)
				assert.Contains(t, page, "data-playground-preview-value")
				assert.Contains(t, page, `data-playground-local-label-value="Run locally"`)
				assert.Contains(t, page, `data-playground-local-title-value="Show the commands that run and test this code (Ctrl/Cmd&#43;Enter)"`)
			} else {
				assert.Contains(t, page, "<span>Run locally</span>")
				assert.Contains(t, page, `title="Show the commands that run and test this code (Ctrl/Cmd&#43;Enter)"`)
				assert.NotContains(t, page, "data-playground-preview-value")
			}
		})
	}

	_, v := New(validDeps()).GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: url.Values{"code": {render}}})
	var buf bytes.Buffer
	require.NoError(t, v.Render(&buf))
	assert.Contains(t, buf.String(), "<span>Run locally</span>", "a shared snippet exists nowhere on the chain")
}

// TestControlsNeedJS checks the playground serves its controls hidden, as
// they only work once the controller connects, and its code shown.
func TestControlsNeedJS(t *testing.T) {
	t.Parallel()

	deps := validDeps()
	deps.Packages = &stubPackages{files: []File{{Name: "foo.gno", Content: "package foo\n"}}}
	h := New(deps)
	_, copyView := h.copyView(context.Background(), "/r/demo/foo", "")
	_, shared := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: url.Values{"code": {"package main\n"}}})
	for name, v := range map[string]*components.View{"copy": copyView, "shared snippet": shared} {
		var buf bytes.Buffer
		require.NoError(t, v.Render(&buf), name)
		page := buf.String()
		for _, hidden := range []string{
			`<div class="b-btns" data-playground-target="js-only" hidden>`,
			`<div class="b-playground-tabs-wrap" data-playground-target="tabs-wrap" hidden>`,
			`<section class="b-result" data-playground-target="js-only" hidden>`,
		} {
			assert.Contains(t, page, hidden, name)
		}
		assert.Contains(t, page, "\npackage ", name)
		files, initial := servedCode(t, page)
		assert.True(t, len(files) > 0 || initial != "", name)
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
// Content-Type or origin, or answered from the cache, costs the reader
// nothing: readers behind one proxy share a bucket, and a popular page
// must not drain it.
func TestEvalLimiter(t *testing.T) {
	t.Parallel()

	client := &stubClient{evalResult: []byte("ok")}
	h := New(Deps{
		Client:   client,
		Packages: &stubPackages{},
		Logger:   discardLogger(),
		Limiter:  &denyLimiter{n: 2},
	})
	handler := h.EvalHandler()

	requests := []struct {
		expr    string
		headers map[string]string
		want    int
	}{
		{"A()", map[string]string{"Content-Type": "text/plain"}, http.StatusUnsupportedMediaType},
		{"A()", map[string]string{"Content-Type": "application/json", "Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"A()", map[string]string{"Content-Type": "application/json"}, http.StatusOK},
		{"B()", map[string]string{"Content-Type": "application/json"}, http.StatusOK},
		// The budget is spent: only cached answers are served.
		{"A()", map[string]string{"Content-Type": "application/json"}, http.StatusOK},
		{"B()", map[string]string{"Content-Type": "application/json"}, http.StatusOK},
		{"C()", map[string]string{"Content-Type": "application/json"}, http.StatusTooManyRequests},
	}
	for i, tc := range requests {
		body := `{"pkg_path":"r/mock/path","expression":"` + tc.expr + `"}`
		req := httptest.NewRequest(http.MethodPost, "/_/api/eval", strings.NewReader(body))
		for k, v := range tc.headers {
			req.Header.Set(k, v)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, tc.want, rr.Code, "request %d: %s", i, tc.expr)
	}
	assert.Equal(t, "B()", client.expr, "C() never reached the node")
}
