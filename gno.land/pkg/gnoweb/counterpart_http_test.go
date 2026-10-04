package gnoweb_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

func serveCounterpart(t *testing.T, client gnoweb.ClientAdapter, target string) *httptest.ResponseRecorder {
	t.Helper()

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		newTestHandlerConfig(t, client),
	)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
	return rr
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
				`<span class="item-label">All by alice</span>`,
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

// A lookup that never answers must not hold the page back, and must be
// cancelled once the page gives up on it.
func TestCounterpart_SlowLookupDoesNotDelayPage(t *testing.T) {
	t.Parallel()

	cancelled := make(chan struct{})
	client := &stubClient{
		realmFunc: func(context.Context, string, string) ([]byte, error) {
			return []byte("hello"), nil
		},
		listPathsFunc: func(ctx context.Context, _ string, _ int) ([]string, error) {
			<-ctx.Done()
			close(cancelled)
			return nil, ctx.Err()
		},
	}

	start := time.Now()
	rr := serveCounterpart(t, client, "/r/alice/golf/game")
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotContains(t, rr.Body.String(), "item--primary")

	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("lookup was not cancelled")
	}
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

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		newTestHandlerConfig(t, client),
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/alice/golf/game", nil)
	req.Header.Set("Accept", "text/markdown")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assert.Zero(t, calls.Load())
}
