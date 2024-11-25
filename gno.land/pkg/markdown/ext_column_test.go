package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func TestExtColumnSimple(t *testing.T) {
	cases := []struct {
		Name     string
		Markdown string
	}{
		{
			"basic", `
===
col1
+++
col2
===
`,
		},
		{
			"left", `
<==
col1
+++
col2
<==
`,
		},
		{
			"right", `
==>
col1
+++
col2
==>
`,
		},
		{
			"middle", `
<==>
col1
+++
col2
<==>
`,
		},
	}

	m := goldmark.New()
	Column.Extend(m)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			source := []byte(tc.Markdown)
			t.Log("input:\n" + strings.TrimSpace(tc.Markdown))

			node := m.Parser().Parse(
				text.NewReader(source),
			)

			var buf bytes.Buffer
			m.Renderer().Render(&buf, []byte(tc.Markdown), node)
			output := buf.Bytes()

			// _, err := html.Parse(bytes.NewReader(output))
			// require.NoError(t, err)

			name := fmt.Sprintf("column_simple_%s.golden", tc.Name)
			file := filepath.Join("testdata", name)

			if *update {
				err := os.WriteFile(file, buf.Bytes(), 0644)
				require.NoError(t, err)
			}

			t.Logf("testfile %q:\n%s", file, buf.String())

			expected, err := os.ReadFile(file)
			require.NoError(t, err)

			assert.Equal(t, expected, buf.Bytes())
		})
	}
}

func TestExtColumn2Simple(t *testing.T) {
	cases := []struct {
		Name     string
		Markdown string
	}{
		{
			"basic", `
---
# Column1

content 1
content 2

---

# Column2

content 1
content 2
---
`,
		},
	}

	m := goldmark.New()
	Column.Extend(m)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			source := []byte(tc.Markdown)
			t.Log("input:\n" + strings.TrimSpace(tc.Markdown))

			node := m.Parser().Parse(
				text.NewReader(source),
			)

			node.Dump(source, 1)

			var buf bytes.Buffer
			m.Renderer().Render(&buf, []byte(tc.Markdown), node)
			output := buf.Bytes()

			// _, err := html.Parse(bytes.NewReader(output))
			// require.NoError(t, err)

			name := fmt.Sprintf("column_simple_%s.golden", tc.Name)
			file := filepath.Join("testdata", name)

			if *update {
				err := os.WriteFile(file, buf.Bytes(), 0644)
				require.NoError(t, err)
			}

			t.Logf("testfile %q:\n%s", file, buf.String())

			expected, err := os.ReadFile(file)
			require.NoError(t, err)

			assert.Equal(t, expected, buf.Bytes())
		})
	}
}
