package markdown

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/extensions"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const (
	goldenRealmTestDir = "golden/realm"
	goldenDocTestDir   = "golden/doc"
)

var (
	update = flag.Bool("update-golden-tests", false, "update golden tests")
	dump   = flag.Bool("dump", false, "dump ast tree after parsing")
)

// createTestExtension creates a test extension with optional options
func createTestExtension(createExt func(...Option) *GnoExtension, opts ...Option) func() *GnoExtension {
	return func() *GnoExtension {
		return createExt(opts...)
	}
}

// testGoldmarkOutputWithExtension tests markdown rendering with a specific extension
func testGoldmarkOutputWithExtension(t *testing.T, nameIn string, input []byte, createExtension func() *GnoExtension) (string, []byte) {
	t.Helper()

	assertExt(t, nameIn, ".md")

	name := nameIn[:len(nameIn)-3] // strip .md
	require.Greater(t, len(name), 0, "txtar file name cannot be empty")

	// Create a test URL for the context
	gnourl, err := weburl.Parse("https://gno.land/r/test")
	require.NoError(t, err)

	// Create parser context with the test URL
	ctxOpts := parser.WithContext(extensions.NewGnoParserContext(gnourl))

	// Use the provided extension
	ext := createExtension()

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

// testDocGoldmarkOutput tests documentation extension
func testDocGoldmarkOutput(t *testing.T, nameIn string, input []byte) (string, []byte) {
	t.Helper()
	return testGoldmarkOutputWithExtension(t, nameIn, input, createTestExtension(NewDocumentationGnoExtension))
}

func TestGeneralExtensions(t *testing.T) {
	gold := NewGoldentTests(testRealmGoldmarkOutput)
	gold.Update = *update
	gold.Recurse = true
	gold.Run(t, goldenRealmTestDir)
}

// testGoldmarkOutput tests realm extension
func testRealmGoldmarkOutput(t *testing.T, nameIn string, input []byte) (string, []byte) {
	t.Helper()
	return testGoldmarkOutputWithExtension(t, nameIn, input,
		createTestExtension(NewRealmGnoExtension, WithImageValidator(func(uri string) bool {
			return !strings.HasPrefix(uri, "https://") // disallow https
		})),
	)
}

func TestDocExtensions(t *testing.T) {
	gold := NewGoldentTests(testDocGoldmarkOutput)
	gold.Update = *update
	gold.Recurse = true
	gold.Run(t, goldenDocTestDir)
}

func assertExt(t *testing.T, filename, ext string) {
	t.Helper()

	require.Truef(t, strings.HasSuffix(filename, ext),
		"expected %q extension for filename %q", ext, filename)
}
