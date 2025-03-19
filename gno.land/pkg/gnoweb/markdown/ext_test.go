package markdown

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
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

	m := goldmark.New()
	GnoExtension.Extend(m)

	// Parse markdown input
	node := m.Parser().Parse(text.NewReader(input))

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
