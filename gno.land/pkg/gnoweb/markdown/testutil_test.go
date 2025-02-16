package markdown

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

const testdataDir = "golden"

var update = flag.Bool("update-golden-file", false, "update golden files")

var makedir = sync.Map{}

func testGoldamarkGoldenOuput(t *testing.T, m goldmark.Markdown, input string) {
	t.Helper()

	// Use txtar for syntaxt highlighting
	name := fmt.Sprintf("%s.golden.txtar", t.Name())
	file := filepath.Join(testdataDir, name)
	testdir := filepath.Dir(file)
	if *update {
		var newOnce sync.Once
		val, _ := makedir.LoadOrStore(testdir, &newOnce)
		val.(*sync.Once).Do(func() {
			err := os.MkdirAll(testdir, 0755)
			require.NoError(t, err)
		})
	}

	var golden bytes.Buffer

	source := []byte(input)
	fmt.Fprint(&golden, "-- input.md --\n")
	fmt.Fprint(&golden, strings.TrimSpace(input)+"\n")

	node := m.Parser().Parse(
		text.NewReader(source),
	)

	// node.Dump(source, 1)

	var buf bytes.Buffer
	m.Renderer().Render(&buf, source, node)
	output := buf.Bytes()

	// Validate html
	_, err := html.Parse(bytes.NewReader(output))
	require.NoError(t, err)

	fmt.Fprint(&golden, "\n-- output.html --\n")
	golden.Write(output)
	golden.WriteRune('\n')

	if *update {
		err := os.WriteFile(file, golden.Bytes(), 0644)
		require.NoError(t, err)
	}

	t.Logf("testfile %s:\n%s", file, golden.String())

	expected, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		require.FailNowf(t, "run with --update-golden-file to generate golden file",
			"%q does't not exist", file,
		)
	}

	require.NoError(t, err)
	assert.Equal(t, expected, golden.Bytes())
}
