package gnoweb

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/stretchr/testify/assert"
)

func TestFormatSource(t *testing.T) {
	// Setup test client with a no-op logger
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	cfg := &HTMLWebClientConfig{
		Domain:      "gno.land",
		ChromaStyle: chromaDefaultStyle,
		ChromaHTMLOptions: []chromahtml.Option{
			chromahtml.WithClasses(true),
			chromahtml.ClassPrefix("chroma-"),
		},
	}
	client := NewHTMLClient(logger, cfg)

	cases := []struct {
		name          string
		fileName      string
		input         string
		expectedPaths []string
		linkable      bool
	}{
		{
			name:          "single import",
			fileName:      "test.gno",
			input:         `import "gno.land/a/b/abc"`,
			expectedPaths: []string{"a/b/abc"},
			linkable:      true,
		},
		{
			name:     "multiple imports",
			fileName: "test.gno",
			input: `import (
				"gno.land/a/b/abc"
				"gno.land/d/e/def"
			)`,
			expectedPaths: []string{"a/b/abc", "d/e/def"},
			linkable:      true,
		},
		{
			name:          "named import",
			fileName:      "test.gno",
			input:         `import name "gno.land/x/y/xyz"`,
			expectedPaths: []string{"x/y/xyz"},
			linkable:      true,
		},
		{
			name:          "empty name import",
			fileName:      "test.gno",
			input:         `import _ "gno.land/a/b/abc"`,
			expectedPaths: []string{"a/b/abc"},
			linkable:      true,
		},
		{
			name:     "multiple imports with name and empty name",
			fileName: "test.gno",
			input: `import (
				name "gno.land/x/y/xyz"
				_ "gno.land/a/b/abc"
			)`,
			expectedPaths: []string{"x/y/xyz", "a/b/abc"},
			linkable:      true,
		},
		{
			name:          "non gno file",
			fileName:      "test.go",
			input:         `import "gno.land/a/b/abc"`,
			expectedPaths: []string{},
			linkable:      false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := client.FormatSource(&buf, tt.fileName, []byte(tt.input))
			assert.NoError(t, err)

			result := buf.String()

			if tt.linkable {
				// Check each expected path has a corresponding link
				for _, path := range tt.expectedPaths {
					expectedLink := fmt.Sprintf(`<a href="/%s$source"`, path)
					assert.Contains(t, result, expectedLink,
						"Should contain link to source for path %s", path)
					assert.Contains(t, result, `class="text-blue-600 hover:underline"`)
				}
			} else {
				assert.NotContains(t, result, `<a href=`,
					"Should not contain any links")
			}

			// Check syntax highlighting is present
			assert.Contains(t, result, `<span class="chroma-`)
		})
	}
}
