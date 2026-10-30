package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestCacheHandler(t *testing.T) {
	t.Parallel()

	h := CacheHandler(`"abc"`, "123", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "missing") {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("asset"))
	}))

	cases := []struct {
		name, url, ifNoneMatch string
		wantStatus             int
		wantCache, wantETag    string
	}{
		{"versioned", "/public/js/index.js?v=123", "", http.StatusOK, immutableCacheControl, `"abc"`},
		{"versioned revalidation", "/public/js/index.js?v=123", `"abc"`, http.StatusNotModified, immutableCacheControl, `"abc"`},
		{"other version", "/public/js/index.js?v=456", "", http.StatusOK, "no-cache", `"abc"`},
		{"empty version", "/public/js/index.js?v=", "", http.StatusOK, "no-cache", `"abc"`},
		{"unversioned import", "/public/js/controller.js", "", http.StatusOK, "no-cache", `"abc"`},
		{"unversioned revalidation", "/public/js/controller.js", `"abc"`, http.StatusNotModified, "no-cache", `"abc"`},
		{"stale copy", "/public/js/controller.js", `"old"`, http.StatusOK, "no-cache", `"abc"`},
		{"unversioned stylesheet", "/public/main.css", "", http.StatusOK, "no-cache", `"abc"`},
		{"content-hashed chunk", "/public/js/chunk-L7IZDLJO.js", "", http.StatusOK, immutableCacheControl, `"abc"`},
		{"woff2 font", "/public/fonts/intervar/Intervar.woff2", "", http.StatusOK, immutableCacheControl, `"abc"`},
		{"woff font", "/public/fonts/roboto/roboto-mono-normal.woff", "", http.StatusOK, immutableCacheControl, `"abc"`},
		{"versioned 404", "/public/js/missing.js?v=123", "", http.StatusNotFound, "no-store", ""},
		{"font 404", "/public/fonts/missing.woff2", "", http.StatusNotFound, "no-store", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			if tc.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tc.ifNoneMatch)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			assert.Equal(t, tc.wantStatus, rr.Code)
			assert.Equal(t, tc.wantCache, rr.Header().Get("Cache-Control"))
			assert.Equal(t, tc.wantETag, rr.Header().Get("ETag"))
		})
	}
}

// TestCacheHandlerFileServer checks the 304 http.FileServer sends itself, for
// an If-None-Match that is not the exact ETag, keeps its headers.
func TestCacheHandlerFileServer(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"controller.js":   {Data: []byte("asset")},
		"chunk-ABCDEF.js": {Data: []byte("asset")},
	}
	h := CacheHandler(`"abc"`, "123", http.FileServerFS(files))

	cases := []struct {
		name, url, ifNoneMatch string
		wantCache              string
	}{
		{"weak", "/controller.js", `W/"abc"`, "no-cache"},
		{"list", "/controller.js", `"old", "abc"`, "no-cache"},
		{"any", "/controller.js", "*", "no-cache"},
		{"weak on a hashed chunk", "/chunk-ABCDEF.js", `W/"abc"`, immutableCacheControl},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			req.Header.Set("If-None-Match", tc.ifNoneMatch)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNotModified, rr.Code)
			assert.Equal(t, tc.wantCache, rr.Header().Get("Cache-Control"))
			assert.Equal(t, `"abc"`, rr.Header().Get("ETag"))
		})
	}
}
