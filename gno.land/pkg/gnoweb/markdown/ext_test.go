package markdown

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const testdataDir = "golden"

var (
	update = flag.Bool("update-golden-tests", false, "update golden tests")
	dump   = flag.Bool("dump", false, "dump ast tree after parsing")
)

func testGoldmarkOutput(t *testing.T, nameIn string, input []byte) (string, []byte) {
	t.Helper()

	assertExt(t, nameIn, ".md")

	name := nameIn[:len(nameIn)-3]
	require.Greater(t, len(name), 0, "txtar file name cannot be empty")

	// Create a test URL for the context
	gnourl, err := weburl.Parse("https://gno.land/r/test")
	require.NoError(t, err)

	// Create parser context with the test URL
	ctxOpts := parser.WithContext(NewGnoParserContext(GnoContext{
		GnoURL: gnourl,
	}))

	ext := NewGnoExtension(WithImageValidator(func(uri string) bool {
		return !strings.HasPrefix(uri, "https://") // disallow https
	}))

	// Create markdown processor with extensions and renderer options
	m := goldmark.New()
	ext.Extend(m)

	// Parse markdown input with context
	node := m.Parser().Parse(text.NewReader(input), ctxOpts)

	// Dump ast to stdout if requested
	if *dump {
		node.Dump(input, 1)
	}

	// Render ast into html
	var html bytes.Buffer
	m.Renderer().Render(&html, input, node)

	return "output.html", html.Bytes()
}

func TestGnoExtension(t *testing.T) {
	gold := NewGoldentTests(testGoldmarkOutput)
	gold.Update = *update
	gold.Recurse = true
	gold.Run(t, testdataDir)
}

func assertExt(t *testing.T, filename, ext string) {
	t.Helper()

	require.Truef(t, strings.HasSuffix(filename, ext),
		"expected %q extension for filename %q", ext, filename)
}
