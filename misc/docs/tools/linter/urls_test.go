package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLintURLs(t *testing.T) {
	t.Parallel()

	var flaky atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/throttled-once", func(w http.ResponseWriter, r *http.Request) {
		if flaky.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
		}
	})
	mux.HandleFunc("/throttled", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	mux.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(time.Second):
		case <-r.Context().Done():
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := urlChecker{
		client:   &http.Client{Timeout: 100 * time.Millisecond},
		attempts: 3,
		backoff:  time.Millisecond,
	}

	tests := []struct {
		path    string
		wantErr error
		wantOut string // "" means no output at all
	}{
		{"/ok", nil, ""},
		{"/missing", errFound404Links, "a.md:3: dead %s: HTTP 404 Not Found\n"},
		{"/throttled-once", nil, ""},
		{"/throttled", nil, "a.md:3: unverified %s: HTTP 429 Too Many Requests (after 3 attempts)\n"},
		{"/unavailable", nil, "a.md:3: unverified %s: HTTP 503 Service Unavailable (after 3 attempts)\n"},
		{"/slow", nil, "a.md:3: unverified %[1]s: Get \"%[1]s\": context deadline exceeded (Client.Timeout exceeded while awaiting headers) (after 3 attempts)\n"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			url := srv.URL + tt.path
			out, err := c.lint(context.Background(), map[string][]urlRef{"a.md": {{url: url, line: 3}}}, true)

			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			}
			if tt.path == "/throttled-once" {
				assert.Equal(t, int32(2), flaky.Load(), "429 then 200 must take exactly one retry")
			}

			if tt.wantOut == "" {
				assert.Empty(t, out)
				return
			}

			assert.Contains(t, out, fmt.Sprintf(tt.wantOut, url))
			assert.Contains(t, out, "1 of 1 remote URLs failed")
		})
	}
}

func TestLintURLsNotFatalWhenWarning(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)

	c := urlChecker{client: srv.Client(), attempts: 1, backoff: time.Millisecond}
	out, err := c.lint(context.Background(), map[string][]urlRef{"a.md": {{url: srv.URL, line: 1}}}, false)

	assert.NoError(t, err)
	assert.Contains(t, out, "a.md:1: dead "+srv.URL+": HTTP 404 Not Found")
}

func TestLintURLsChecksEachURLOnceAndBoundsConcurrency(t *testing.T) {
	t.Parallel()

	var hits, inFlight, peak atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		n := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	refs := map[string][]urlRef{}
	for i := range 40 {
		u := fmt.Sprintf("%s/%d", srv.URL, i)
		refs["a.md"] = append(refs["a.md"], urlRef{url: u, line: i + 1})
		refs["b.md"] = append(refs["b.md"], urlRef{url: u, line: 1})
	}

	c := urlChecker{client: srv.Client(), attempts: 1, backoff: time.Millisecond}
	out, err := c.lint(context.Background(), refs, true)

	require.ErrorIs(t, err, errFound404Links)
	assert.Equal(t, int32(40), hits.Load(), "a URL cited twice is fetched once")
	assert.LessOrEqual(t, peak.Load(), int32(maxConcurrentChecks))
	assert.Contains(t, out, "a.md:1: dead "+srv.URL+"/0: HTTP 404 Not Found\n")
	assert.Contains(t, out, "b.md:1: dead "+srv.URL+"/0: HTTP 404 Not Found\n")
	assert.Contains(t, out, "40 of 40 remote URLs failed: 40 dead, 0 unverified")
}

func TestLintRelativeLink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	tests := []struct {
		name    string
		link    string
		wantErr error
		wantOut string
	}{
		{"exists", "./other.md", nil, "Lint complete, no issues found."},
		{"missing", "./gone.md", errFoundLintItems, "./gone.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sub := filepath.Join(dir, tt.name)
			require.NoError(t, os.MkdirAll(sub, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(sub, "other.md"), []byte("# other\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(sub, "index.md"), []byte("see [it]("+tt.link+")\n"), 0o644))

			out, err := execLint(&cfg{docsPath: sub, treatUrlsAsErr: true}, context.Background())

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Contains(t, out, tt.wantOut)
		})
	}
}
