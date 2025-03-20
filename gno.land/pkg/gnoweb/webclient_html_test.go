package gnoweb

import (
	"net/url"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test URL
			parsedURL, err := url.Parse(tt.url)
			assert.NoError(t, err)
			gnourl, err := weburl.ParseGnoURL(parsedURL)
			assert.NoError(t, err)

			// Create test document with a link
			doc := ast.NewDocument()
			link := ast.NewLink()
			link.Destination = []byte(tt.link)
			doc.AppendChild(doc, link)

			// Transform URLs
			TransformRelArgsURL(doc, gnourl)

			// Verify transformation
			assert.Equal(t, tt.expected, string(link.Destination))
		})
	}
}
