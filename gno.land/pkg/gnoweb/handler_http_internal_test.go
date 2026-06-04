package gnoweb

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalPathURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{
			name:     "plain trailing slash",
			target:   "/r/demo/foo/",
			expected: "/r/demo/foo",
		},
		{
			name:     "web query suffix",
			target:   "/r/demo/foo/$source&file=render.gno",
			expected: "/r/demo/foo$source&file=render.gno",
		},
		{
			name:     "path args with query",
			target:   "/r/demo/foo/:bob?arg1=val1&arg2=val2",
			expected: "/r/demo/foo:bob?arg1=val1&arg2=val2",
		},
		{
			name:     "query only",
			target:   "/r/demo/foo/?arg=value",
			expected: "/r/demo/foo?arg=value",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", tc.target, nil)
			assert.Equal(t, tc.expected, canonicalPathURL(req))
		})
	}
}
