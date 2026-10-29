package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	const immutable = "public, max-age=31536000, immutable"
	cases := []struct {
		name, url, ifNoneMatch string
		wantStatus             int
		wantCache, wantETag    string
	}{
		{"versioned", "/public/js/index.js?v=123", "", http.StatusOK, immutable, `"abc"`},
		{"versioned revalidation", "/public/js/index.js?v=123", `"abc"`, http.StatusNotModified, immutable, `"abc"`},
		{"other version", "/public/js/index.js?v=456", "", http.StatusOK, "no-cache", `"abc"`},
		{"empty version", "/public/js/index.js?v=", "", http.StatusOK, "no-cache", `"abc"`},
		{"unversioned import", "/public/js/controller.js", "", http.StatusOK, "no-cache", `"abc"`},
		{"unversioned revalidation", "/public/js/controller.js", `"abc"`, http.StatusNotModified, "no-cache", `"abc"`},
		{"stale copy", "/public/js/controller.js", `"old"`, http.StatusOK, "no-cache", `"abc"`},
		{"unversioned stylesheet", "/public/main.css", "", http.StatusOK, "no-cache", `"abc"`},
		{"content-hashed chunk", "/public/js/chunk-L7IZDLJO.js", "", http.StatusOK, immutable, `"abc"`},
		{"woff2 font", "/public/fonts/intervar/Intervar.woff2", "", http.StatusOK, immutable, `"abc"`},
		{"woff font", "/public/fonts/roboto/roboto-mono-normal.woff", "", http.StatusOK, immutable, `"abc"`},
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
