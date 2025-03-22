package markdown

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
)

type GenFunc func(t *testing.T, nameIn string, input []byte) (nameOut string, output []byte)

type GoldenTests struct {
	Recurse      bool
	Update       bool
	GenerateFunc GenFunc
}

func NewGoldentTests(exec GenFunc) *GoldenTests {
	return &GoldenTests{
		Recurse:      true,
		GenerateFunc: exec,
	}
}

func (g *GoldenTests) Run(t *testing.T, dir string) {
	t.Helper()

	files := []string{}
	shouldSkipDir := filepath.SkipDir
	if g.Recurse {
		shouldSkipDir = nil
	}

	// Construct paths list
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return shouldSkipDir
		}

		if ext := filepath.Ext(info.Name()); ext == ".txtar" || ext == ".txt" {
			files = append(files, path)
		}

		return nil
	})

	for _, file := range files {
		name, found := strings.CutPrefix(file, filepath.Clean(dir)+"/")
		// Cleanup name

		require.True(t, found)
		name, _ = strings.CutSuffix(name, filepath.Ext(name))

		// Run individual test by file
		t.Run(name, func(t *testing.T) {
			archive, err := txtar.ParseFile(file)
			require.NoError(t, err)

			var input, expected txtar.File
			switch len(archive.Files) {
			case 0:
				require.Fail(t, "empty txtar file", "path", file)
			case 1:
				// No outout has been generated yet
				input = archive.Files[0]
			case 2:
				input = archive.Files[0]
				expected = archive.Files[1]
			default:
				require.Fail(t, "txtar should be composed with two file, one input and one output")
			}

			var output txtar.File
			output.Name, output.Data = g.GenerateFunc(t, input.Name, input.Data)

			t.Logf("input - %q:\n%s", input.Name, string(input.Data))

			if g.Update {
				// Update the result
				expected = output
				archive.Files = []txtar.File{input, expected}

				err := os.WriteFile(file, txtar.Format(archive), 0o666)
				require.NoError(t, err, "cannot update txtar file")
				t.Logf("%q updated", file)
			}

			t.Logf("output - %q:\n%s", output.Name, string(output.Data))

			if len(archive.Files) == 1 {
				// Nothing expected, log generated output and
				// mark the test as fail
				require.Fail(t, "file need to be updated with `go test -update-golden-files`")
			}

			// Ultimatly compare generated output with expected output
			require.Equal(t, string(expected.Data), string(output.Data))
		})
	}
}
