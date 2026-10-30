package playground

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
// <pre> of the editor that names a file, as a browser parses it.
func servedCode(t *testing.T, page string) (files []File) {
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
		if target, _ := attr(n, "data-playground-target"); target == "file" {
			name, _ := attr(n, "data-playground-name-value")
			files = append(files, File{Name: name, Content: text(n)})
		}
	}
	return files
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

// shareCode is the ?code= value Share writes for files: the base64 of raw
// DEFLATE of their JSON list.
func shareCode(t *testing.T, files []File) string {
	t.Helper()

	b, err := json.Marshal(files)
	require.NoError(t, err)
	return deflateBase64(t, b)
}

// TestGetPlaygroundViewCode covers the "code" shared-snippet paths. Each
// code holds a sentinel the default code does not, so a fallback to the
// default never passes for a decoded snippet.
func TestGetPlaygroundViewCode(t *testing.T) {
	t.Parallel()

	const sentinel = "package sentinel_42\n"
	require.NotContains(t, defaultCode, "sentinel_42")
	main := func(code string) []File { return []File{{Name: "main.gno", Content: code}} }
	def := main(defaultCode)
	one := main(sentinel)
	z := shareCode(t, one)
	share := func(files []File) url.Values { return url.Values{"code": {shareCode(t, files)}, "z": {""}} }
	zcode := func(b []byte) url.Values { return url.Values{"code": {deflateBase64(t, b)}, "z": {""}} }

	// Lines shaped like the separator the text format had, and blank lines
	// at either edge: each file comes back as it went.
	separators := []File{
		{Name: "a.gno", Content: "\n\npackage sentinel_42\n\n// --- b.gno ---\nfunc A() {}\n// --- Base set ---\n\n\n"},
		{Name: "b.gno", Content: "// --- a.gno ---\npackage sentinel_42"},
		{Name: "c.gno", Content: ""},
		{Name: "gnomod.toml", Content: "module = \"sentinel_42\"\n"},
	}
	unicode := []File{{Name: "u.gno", Content: "package sentinel_42 // héllo 世界 🎉 \"\\\" <b>&amp;</b> \u2028\n"}}
	many := make([]File, 60)
	for i := range many {
		many[i] = File{Name: fmt.Sprintf("f%d.gno", i), Content: fmt.Sprintf("package sentinel_42\n\nvar N%d = %d\n", i, i)}
	}
	// A JSON list of exactly n bytes, one file of "a"s.
	sized := func(n int) []File {
		empty, err := json.Marshal([]File{{Name: "a.gno"}})
		require.NoError(t, err)
		return []File{{Name: "a.gno", Content: strings.Repeat("a", n-len(empty))}}
	}

	cases := []struct {
		name    string
		query   url.Values
		want    []File
		refused error // what the page says, nil for nothing
	}{
		{"no code", url.Values{}, def, nil},
		{"z and no code", url.Values{"z": {""}}, def, nil},
		{"z and an empty code", url.Values{"code": {""}, "z": {""}}, def, nil},
		{"plain text", url.Values{"code": {sentinel}}, one, nil},
		// Without &z, code is one main.gno, whatever its lines look like.
		{"plain text with separator lines", url.Values{"code": {"// --- a.gno ---\n" + sentinel + "// --- b.gno ---\n"}}, main("// --- a.gno ---\n" + sentinel + "// --- b.gno ---\n"), nil},
		{"plain text that parses as base64", url.Values{"code": {"test"}}, main("test"), nil},
		{"base64 without z stays text", url.Values{"code": {"cGFja2FnZSBtYWluCg=="}}, main("cGFja2FnZSBtYWluCg=="), nil},
		// One meaning per form: without &z nothing is decoded, not even
		// what Share would have written with it.
		{"shared without z stays text", url.Values{"code": {z}}, main(z), nil},
		{"shared", url.Values{"code": {z}, "z": {""}}, one, nil},
		{"shared separator lines and blank edges", share(separators), separators, nil},
		{"shared unicode and markup", share(unicode), unicode, nil},
		{"shared many files", share(many), many, nil},
		{"shared, cut short", url.Values{"code": {z[:len(z)-1]}, "z": {""}}, def, errSharedInvalid},
		{"z on base64 that is no deflate", url.Values{"code": {"dGVzdA=="}, "z": {""}}, def, errSharedInvalid},
		{"z on text that is no base64", url.Values{"code": {"package main"}, "z": {""}}, def, errSharedInvalid},
		// The text format Share wrote before: no list, so refused.
		{"z on the text format", zcode([]byte("// --- a.gno ---\n" + sentinel)), def, errSharedInvalid},
		{"z on a JSON object", zcode([]byte(`{"name":"a.gno","content":""}`)), def, errSharedInvalid},
		{"z on an empty list", zcode([]byte(`[]`)), def, errSharedInvalid},
		{"z on null", zcode([]byte(`null`)), def, errSharedInvalid},
		{"z on a list with trailing data", zcode([]byte(`[{"name":"a.gno","content":""}]x`)), def, errSharedInvalid},
		{"z on a file that is no object", zcode([]byte(`["a.gno"]`)), def, errSharedInvalid},
		// Names from isValidFileName's allowlist, each once.
		{"a path in a name", share([]File{{Name: "../x.gno", Content: sentinel}}), def, errSharedName},
		{"no name", share([]File{{Content: sentinel}}), def, errSharedName},
		{"a name that is no source", share([]File{{Name: "README.md", Content: sentinel}}), def, errSharedName},
		{"a name twice", share([]File{{Name: "a.gno", Content: sentinel}, {Name: "a.gno"}}), def, errSharedName},
		// The cap is written out, not read from the constant, so raising it
		// fails here: 64 KiB is what a link Share writes can hold.
		{"shared up to 64 KiB", share(sized(64 << 10)), sized(64 << 10), nil},
		// A few bytes of a link must not expand into a page of megabytes.
		{"shared past 64 KiB", share(sized(64<<10 + 1)), def, errSharedTooLarge},
		{"a link-sized megabyte", zcode(bytes.Repeat([]byte("a"), 1<<20)), def, errSharedTooLarge},
		{"longer than Share writes", url.Values{"code": {strings.Repeat("a", maxSharedCodeLength+1)}}, def, errSharedTooLong},
		{"shared, longer than Share writes", url.Values{"code": {z + strings.Repeat("A", maxSharedCodeLength)}, "z": {""}}, def, errSharedTooLong},
	}

	h := New(validDeps())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: tc.query})
			assert.Equal(t, http.StatusOK, status)
			data := extractPlaygroundViewData(t, v)
			assert.Equal(t, tc.want, data.Files)
			assert.Empty(t, data.From, "a share is no copy")
			assert.False(t, data.Preview)
			// Served once, as text, one <pre> per file, which the
			// controller reads back as is.
			var buf bytes.Buffer
			require.NoError(t, v.Render(&buf))
			assert.Equal(t, tc.want, servedCode(t, buf.String()))
			assert.NotContains(t, buf.String(), "<textarea", "the code is in the page once")
			// A refused link says so: the default code alone would read
			// as the link's.
			const notice = "This link's code was not opened"
			if tc.refused == nil {
				assert.NotContains(t, buf.String(), notice)
				return
			}
			assert.Contains(t, buf.String(), notice+": "+template.HTMLEscapeString(tc.refused.Error())+". The default code is shown instead.")
		})
	}

	// A CRLF file decodes as it is; the page serves it as the browser and
	// the editor read it, with LF line ends.
	crlf := []File{{Name: "crlf.gno", Content: "package sentinel_42\r\n\r\nfunc F() {}\r\n"}}
	_, v := h.GetPlaygroundView(context.Background(), &weburl.GnoURL{Query: share(crlf)})
	assert.Equal(t, crlf, extractPlaygroundViewData(t, v).Files)
	var buf bytes.Buffer
	require.NoError(t, v.Render(&buf))
	assert.Equal(t, []File{{Name: "crlf.gno", Content: "package sentinel_42\n\nfunc F() {}\n"}}, servedCode(t, buf.String()))
}

// TestIsValidFileName holds the cases of playground-files.test.mts's
// isValidFileName test: a share opens only what Add file would take.
func TestIsValidFileName(t *testing.T) {
	t.Parallel()

	for name, want := range map[string]bool{
		"main.gno":                       true,
		"gnomod.toml":                    true,
		"a_test.gno":                     true,
		"Z-1.v2.gno":                     true,
		strings.Repeat("x", 96) + ".gno": true,  // 100 bytes
		strings.Repeat("x", 97) + ".gno": false, // 101 bytes, cut by a ustar header
		"$(touch PWNED).gno":             false,
		"a;b.gno":                        false,
		"a b.gno":                        false,
		"../x.gno":                       false,
		"x/y.gno":                        false,
		".hidden.gno":                    false,
		"-rf.gno":                        false,
		"é.gno":                          false,
		"README.md":                      false,
		"Base set":                       false,
		".gno":                           false,
		"":                               false,
		"a.gno\n":                        false,
	} {
		assert.Equal(t, want, isValidFileName(name), "%q", name)
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
		assert.Equal(t, "b.gno", data.DefaultFile)
		assert.Equal(t, "r/demo/foo", data.From)
		assert.Equal(t, deps.Domain, data.Domain)
		assert.Equal(t, "README.md", pkgs.files[1].Name, "the fetcher's slice is left as it was")

		// The page serves each file once, as text under its name, which a
		// reader without JS reads and the controller reads back as is.
		var buf bytes.Buffer
		require.NoError(t, v.Render(&buf))
		page := buf.String()
		assert.Equal(t, data.Files, servedCode(t, page))
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
		got := servedCode(t, buf.String())
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
		assert.NotEmpty(t, servedCode(t, page), name)
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

	deflate := func(b []byte) []byte {
		var buf bytes.Buffer
		zw, _ := flate.NewWriter(&buf, flate.BestCompression)
		_, _ = zw.Write(b)
		require.NoError(t, zw.Close())
		return buf.Bytes()
	}

	got, err := decodeCompressedCode(deflate([]byte("some gno source")))
	require.NoError(t, err)
	assert.Equal(t, "some gno source", string(got))

	_, err = decodeCompressedCode(deflate(bytes.Repeat([]byte{0}, maxDecompressedCodeSize+1)))
	assert.ErrorIs(t, err, errSharedTooLarge, "payload exceeding the ceiling must be rejected")

	_, err = decodeCompressedCode([]byte("not deflate data"))
	assert.ErrorIs(t, err, errSharedInvalid)
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

// TestShareCapsMatch checks Share refuses exactly the links the server
// would not open: both sides hold the caps, so a change to one alone
// would let Share write links that open as the default code.
func TestShareCapsMatch(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("frontend/playground-files.ts")
	require.NoError(t, err)
	assert.Contains(t, string(src), fmt.Sprintf("MAX_SHARE_URL_LENGTH = %d_%03d;", maxSharedCodeLength/1000, maxSharedCodeLength%1000))
	assert.Contains(t, string(src), fmt.Sprintf("MAX_SHARE_CODE_BYTES = %d * 1024;", maxDecompressedCodeSize/1024))
}
