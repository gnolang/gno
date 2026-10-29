package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheHandler(t *testing.T) {
	t.Parallel()

	h := CacheHandler(`"abc"`, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("asset"))
	}))

	cases := []struct {
		name, url, ifNoneMatch string
		wantStatus             int
		wantCache              string
	}{
		{"versioned", "/public/js/index.js?v=123", "", http.StatusOK, "public, max-age=31536000, immutable"},
		{"unversioned import", "/public/js/controller.js", "", http.StatusOK, "no-cache"},
		{"unversioned revalidation", "/public/js/controller.js", `"abc"`, http.StatusNotModified, "no-cache"},
		{"stale copy", "/public/js/controller.js", `"old"`, http.StatusOK, "no-cache"},
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
			assert.Equal(t, `"abc"`, rr.Header().Get("ETag"))
		})
	}
}
