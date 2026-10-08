package gnoweb_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var renderFuncs = []*doc.JSONFunc{{
	Name:    "Render",
	Params:  []*doc.JSONField{{Name: "path", Type: "string"}},
	Results: []*doc.JSONField{{Type: "string"}},
}}

func golfPackages() []*gnoweb.MockPackage {
	return []*gnoweb.MockPackage{
		{
			Path:      "/r/alice/golf/game",
			Files:     map[string]string{"game.gno": "package game"},
			Functions: renderFuncs,
		},
		{Path: "/p/alice/golf/course", Files: map[string]string{"course.gno": "package course"}},
		{Path: "/p/alice/golf/physics", Files: map[string]string{"physics.gno": "package physics"}},
		{Path: "/p/alice/golfer", Files: map[string]string{"golfer.gno": "package golfer"}},
		{
			Path:      "/r/bob/solo",
			Files:     map[string]string{"solo.gno": "package solo"},
			Functions: renderFuncs,
		},
	}
}

func newCounterpartHandler(t *testing.T, client gnoweb.ClientAdapter) http.Handler {
	t.Helper()

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		newTestHandlerConfig(t, client),
	)
	require.NoError(t, err)
	return handler
}

func serve(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func serveCounterpart(t *testing.T, client gnoweb.ClientAdapter, target string) *httptest.ResponseRecorder {
	t.Helper()
	return serve(newCounterpartHandler(t, client), httptest.NewRequest(http.MethodGet, target, nil))
}

func TestCounterpart_HeaderLink(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		target  string
		want    []string
		notWant []string
	}{
		{
			name:   "realm links the project's packages",
			target: "/r/alice/golf/game",
			want: []string{
				`popovertarget="kind-switch-menu"`,
				`<a href="/p/alice/golf" class="item item--primary">`,
				`<span class="item-label">2 matching packages</span>`,
				`<span class="item-path">/p/alice/golf</span>`,
				`<use href="#ico-pure"></use>`,
				`<a href="/r/" class="item item--inline">`,
				`<a href="/u/alice" class="item item--inline">`,
				`<span class="item-label">All in alice</span>`,
			},
			notWant: []string{"/p/alice/golfer"},
		},
		{
			name:   "source tab keeps the link",
			target: "/r/alice/golf/game$source",
			want:   []string{`<a href="/p/alice/golf" class="item item--primary">`},
		},
		{
			name:   "state page keeps the link",
			target: "/r/alice/golf/game$state",
			want:   []string{`<a href="/p/alice/golf" class="item item--primary">`},
		},
		{
			name:   "package links its only realm directly",
			target: "/p/alice/golf/physics",
			want: []string{
				`<a href="/r/alice/golf/game" class="item item--primary">`,
				`<span class="item-label">Matching realm</span>`,
				`<span class="item-label">All packages</span>`,
			},
		},
		{
			name:   "project listing links the other side's listing",
			target: "/r/alice/golf/",
			want:   []string{`<a href="/p/alice/golf" class="item item--primary">`},
		},
		{
			name:   "nothing on the other side keeps the menu without the switch",
			target: "/r/bob/solo",
			want: []string{
				`popovertarget="kind-switch-menu"`,
				`<a href="/r/" class="item item--inline">`,
				`<a href="/u/bob" class="item item--inline">`,
			},
			notWant: []string{"item--primary"},
		},
		{
			name:    "kind listing has no namespace, no menu",
			target:  "/r/",
			notWant: []string{"kind-switch"},
		},
		{
			name:    "user page has no counterpart",
			target:  "/u/alice",
			notWant: []string{"kind-switch"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := serveCounterpart(t, gnoweb.NewMockClient(golfPackages()...), tc.target)
			body := rr.Body.String()
			for _, s := range tc.want {
				assert.Contains(t, body, s)
			}
			for _, s := range tc.notWant {
				assert.NotContains(t, body, s)
			}
		})
	}
}

func TestCounterpart_LookupFailureKeepsPage(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		realmFunc: func(context.Context, string, string) ([]byte, error) {
			return []byte("hello"), nil
		},
		listPathsFunc: func(context.Context, string, int) ([]string, error) {
			return nil, errors.New("node down")
		},
	}

	rr := serveCounterpart(t, client, "/r/alice/golf/game")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "hello")
	assert.NotContains(t, rr.Body.String(), "item--primary")
}

// A slow lookup must not hold the page back, and its answer, once in, must
// serve the next page from the cache.
func TestCounterpart_SlowLookupFillsCache(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	answered := make(chan struct{})
	client := &stubClient{
		realmFunc: func(context.Context, string, string) ([]byte, error) {
			return []byte("hello"), nil
		},
		listPathsFunc: func(ctx context.Context, _ string, _ int) ([]string, error) {
			calls.Add(1)
			defer close(answered)
			select {
			case <-time.After(time.Second):
				return []string{"/p/alice/golf/game"}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}
	handler := newCounterpartHandler(t, client)
	get := func() *httptest.ResponseRecorder {
		return serve(handler, httptest.NewRequest(http.MethodGet, "/r/alice/golf/game", nil))
	}

	start := time.Now()
	rr := get()
	assert.Less(t, time.Since(start), 900*time.Millisecond)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotContains(t, rr.Body.String(), "item--primary")

	select {
	case <-answered:
	case <-time.After(3 * time.Second):
		t.Fatal("lookup did not finish")
	}
	assert.Contains(t, get().Body.String(), `<a href="/p/alice/golf/game" class="item item--primary">`)
	assert.Equal(t, int32(1), calls.Load())
}

// State fragments and JSON render no header, so they must not pay for a lookup.
func TestCounterpart_StateAPISkipsLookup(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	client := &stubClient{
		listPathsFunc: func(context.Context, string, int) ([]string, error) {
			calls.Add(1)
			return nil, nil
		},
	}

	for _, target := range []string{
		"/r/alice/golf/game$state&json",
		"/r/alice/golf/game$state&frag=tree",
	} {
		serveCounterpart(t, client, target)
	}
	assert.Zero(t, calls.Load())
}

// A markdown answer (Accept: text/markdown) renders no header either.
func TestCounterpart_MarkdownAnswerSkipsLookup(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	client := &stubClient{
		listPathsFunc: func(context.Context, string, int) ([]string, error) {
			calls.Add(1)
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/r/alice/golf/game", nil)
	req.Header.Set("Accept", "text/markdown")
	serve(newCounterpartHandler(t, client), req)
	assert.Zero(t, calls.Load())
}

var rePrimaryLink = regexp.MustCompile(`<a href="([^"]*)" class="item item--primary">\s*<svg[^>]*><use[^>]*></use></svg>\s*<span class="item-label">([^<]*)</span>\s*<span class="item-path">([^<]*)</span>`)

// The "N matching" line must open a page that lists exactly those N paths. A
// directory that is itself a package or realm opens that one package, so the
// line then points at its overview's Directories section, which lists its
// direct children when two or more match, or names the package alone.
// Reported by davd-gzl on #6262.
func TestCounterpart_LinkOpensWhatItCounts(t *testing.T) {
	t.Parallel()

	realm := func(path string) *gnoweb.MockPackage {
		return &gnoweb.MockPackage{Path: path, Files: map[string]string{"r.gno": "package r"}, Functions: renderFuncs}
	}
	pure := func(path string) *gnoweb.MockPackage {
		return &gnoweb.MockPackage{Path: path, Files: map[string]string{"p.gno": "package p"}}
	}

	cases := []struct {
		name   string
		pkgs   []*gnoweb.MockPackage
		page   string
		href   string
		label  string
		listed []string // what the opened page must show, one per counted path
	}{
		{
			name:   "twin's directory is a realm (/r/tests/vm shape)",
			pkgs:   []*gnoweb.MockPackage{realm("/r/tests/vm"), realm("/r/tests/vm/crossrealm"), realm("/r/tests/vm/subtests"), pure("/p/tests/vm/crossrealm")},
			page:   "/p/tests/vm/crossrealm",
			href:   "/r/tests/vm$source#subpackages",
			label:  "2 matching realms",
			listed: []string{"/r/tests/vm/crossrealm", "/r/tests/vm/subtests"},
		},
		{
			name:   "twin's directory is a package",
			pkgs:   []*gnoweb.MockPackage{pure("/p/alice/golf"), pure("/p/alice/golf/v1"), pure("/p/alice/golf/v2"), realm("/r/alice/golf/v1")},
			page:   "/r/alice/golf/v1",
			href:   "/p/alice/golf$source#subpackages",
			label:  "2 matching packages",
			listed: []string{"/p/alice/golf/v1", "/p/alice/golf/v2"},
		},
		{
			name:   "no twin, a realm above it has matching children (/r/tests/vm shape)",
			pkgs:   []*gnoweb.MockPackage{realm("/r/tests/vm"), realm("/r/tests/vm/crossrealm"), realm("/r/tests/vm/subtests"), pure("/p/tests/vm/foo")},
			page:   "/p/tests/vm/foo",
			href:   "/r/tests/vm$source#subpackages",
			label:  "2 matching realms",
			listed: []string{"/r/tests/vm/crossrealm", "/r/tests/vm/subtests"},
		},
		{
			name:  "no twin, project root is a realm (/r/gov/dao shape)",
			pkgs:  []*gnoweb.MockPackage{realm("/r/gov/dao"), realm("/r/gov/dao/impl/v0"), realm("/r/gov/dao/init/v0"), pure("/p/gov/dao/utils")},
			page:  "/p/gov/dao/utils",
			href:  "/r/gov/dao",
			label: "Matching realm",
		},
		{
			name:   "twin's listing counts the whole subtree it shows",
			pkgs:   []*gnoweb.MockPackage{pure("/p/alice/golf/v0"), pure("/p/alice/golf/v2"), pure("/p/alice/golf/ui/board"), realm("/r/alice/golf/v0")},
			page:   "/r/alice/golf/v0",
			href:   "/p/alice/golf",
			label:  "3 matching packages",
			listed: []string{"/p/alice/golf/v0", "/p/alice/golf/v2", "/p/alice/golf/ui/board"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newCounterpartHandler(t, gnoweb.NewMockClient(tc.pkgs...))
			body := serve(h, httptest.NewRequest(http.MethodGet, tc.page, nil)).Body.String()
			m := rePrimaryLink.FindStringSubmatch(body)
			require.NotNil(t, m, "no switch link on %s", tc.page)
			assert.Equal(t, tc.href, m[1])
			assert.Equal(t, tc.label, m[2])
			// The grey path line shows the plain path, without the tab or anchor.
			plain, _, _ := strings.Cut(tc.href, "$")
			assert.Equal(t, plain, m[3])

			target, fragment, _ := strings.Cut(tc.href, "#")
			rr := serve(h, httptest.NewRequest(http.MethodGet, target, nil))
			require.Equal(t, http.StatusOK, rr.Code)
			opened := rr.Body.String()
			if fragment != "" {
				assert.True(t, strings.Contains(opened, `id="`+fragment+`"`), "%s has no #%s", target, fragment)
			}
			for _, p := range tc.listed {
				assert.True(t, strings.Contains(opened, `href="`+p+`"`), "%s does not list %s", target, p)
			}
		})
	}
}
