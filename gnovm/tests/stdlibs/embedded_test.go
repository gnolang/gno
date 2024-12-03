package stdlibs

import (
	"embed"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbedTree(t *testing.T) {
	actualEmbedTree := dumpEmbedFS(t, embeddedSources, 0, ".")
	require.Equal(t, expectedEmbedTree, "\n"+actualEmbedTree)
}

func dumpEmbedFS(t *testing.T, efs embed.FS, level int, p string) string {
	t.Helper()

	s := ""

	dir, err := efs.ReadDir(p)
	require.NoError(t, err)

	for _, entry := range dir {
		s += fmt.Sprintf("%s%s\n", strings.Repeat("  ", level), entry.Name())
		if entry.IsDir() {
			s += dumpEmbedFS(t, efs, level+1, path.Join(p, entry.Name()))
		}
	}

	return s
}

const expectedEmbedTree = `
std
  frame_testing.gno
  std.gno
  std.go
testing
  native_testing.gno
  native_testing.go
unicode
  natives.gno
  natives.go
`
