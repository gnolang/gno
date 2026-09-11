package components

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSourceView_ExposesCopyTarget guards the header Copy button: both the
// rendered-markdown README view and the code view must carry data-copy-target
// so the button has an element to read (regression: README had none, PR #5572).
func TestSourceView_ExposesCopyTarget(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{"readme", "README.md", "<md-renderer>hello</md-renderer>"},
		{"code", "foo.gno", "<pre>code</pre>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data := SourceData{
				PkgPath:    "/r/demo/foo",
				Files:      []string{tc.fileName},
				FileName:   tc.fileName,
				FileSource: rawHTMLComponent(tc.source),
			}
			var buf bytes.Buffer
			require.NoError(t, SourceView(data).Render(&buf))
			require.Contains(t, buf.String(), `data-copy-target="source-code"`,
				"source view must expose the copy target for the header Copy button")
		})
	}
}

// TestSourceView_LinksBackToOverview guards the sidebar affordance that lets a
// reader return from a single file to the package overview.
func TestSourceView_LinksBackToOverview(t *testing.T) {
	t.Parallel()
	data := SourceData{
		PkgPath:    "/r/demo/foo",
		Files:      []string{"foo.gno"},
		FileName:   "foo.gno",
		FileSource: rawHTMLComponent("<pre>code</pre>"),
	}
	var buf bytes.Buffer
	require.NoError(t, SourceView(data).Render(&buf))
	require.Contains(t, buf.String(), `href="/r/demo/foo$source"`,
		"source view sidebar must link back to the package overview")
}
