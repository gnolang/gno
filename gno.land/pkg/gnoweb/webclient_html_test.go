package gnoweb

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/ast"
)

func TestTransformRelArgsURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		link     string
		expected string
	}{
		{
			name:     "basic realm link",
			url:      "http://gno.land/r/demo",
			link:     ":hello",
			expected: "/r/demo:hello",
		},
		{
			name:     "realm link with multiple segments",
			url:      "http://gno.land/r/gov/dao/proxy",
			link:     ":2/votes",
			expected: "/r/gov/dao/proxy:2/votes",
		},
		{
			name:     "realm link with query parameters",
			url:      "http://gno.land/r/demo?help=true",
			link:     ":hello",
			expected: "/r/demo:hello",
		},
		{
			name:     "non-special link should not be transformed",
			url:      "http://gno.land/r/demo",
			link:     "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "realm link with existing args",
			url:      "http://gno.land/r/gov/dao/proxy:1/init",
			link:     ":2/votes",
			expected: "/r/gov/dao/proxy:2/votes",
		},
		{
			name:     "absolute gno.land URL",
			url:      "http://gno.land/r/demo",
			link:     "gno.land/r/demo/hello",
			expected: "gno.land/r/demo/hello",
		},
		{
			name:     "URL with web query",
			url:      "http://gno.land/r/demo",
			link:     ":help?section=getting-started",
			expected: "/r/demo:help?section=getting-started",
		},
		{
			name:     "non-special link with static path",
			url:      "http://gno.land/r/demo",
			link:     "/static/docs.pdf",
			expected: "/static/docs.pdf",
		},
		{
			name:     "empty URL should not transform",
			url:      "",
			link:     ":hello",
			expected: ":hello",
		},
		{
			name:     "realm with existing args and new args",
			url:      "http://gno.land/r/test:bla/1",
			link:     ":2/votes",
			expected: "/r/test:2/votes",
		},
		{
			name:     "realm with multiple existing args",
			url:      "http://gno.land/r/test:arg1/1:arg2/2",
			link:     ":3/votes",
			expected: "/r/test:3/votes",
		},
		{
			name:     "realm with existing args and query",
			url:      "http://gno.land/r/test:arg1/1?help=true",
			link:     ":2/votes",
			expected: "/r/test:2/votes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse URL and extract pkgPath (for testing purposes - simulating the pkgPath)
			var pkgPath string
			if tt.url != "" {
				parsedURL, err := url.Parse(tt.url)
				assert.NoError(t, err)
				// Extract the base path without any args or queries
				path := parsedURL.Path
				if idx := strings.Index(path, ":"); idx != -1 {
					path = path[:idx]
				}
				pkgPath = strings.TrimPrefix(path, "/")
			}

			// Create test document with a link
			doc := ast.NewDocument()
			link := ast.NewLink()
			link.Destination = []byte(tt.link)
			doc.AppendChild(doc, link)

			// Transform URLs
			TransformRelArgsURL(doc, pkgPath)

			// Verify transformation
			assert.Equal(t, tt.expected, string(link.Destination))
		})
	}
}
