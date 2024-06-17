// For convenient purposes, more tests have been created using `testscripts` format and are located in `gnovm/cmd/testdata/gno_fmt/` folder

package gnoimports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatImportFromSource(t *testing.T) {
	mockResolver := NewMockResolver()

	// Add packages to the MockResolver
	mockResolver.AddPackage(&Package{
		Path: "example.com/mypkg",
		Name: "mypkg",
	})

	processor := NewProcessor(mockResolver)
	sourceCode := `package main

func main() {
	str := "hello, world"
	mypkg.MyFunc(str)
}`

	expectedOutput := `package main

import "example.com/mypkg"

func main() {
	str := "hello, world"
	mypkg.MyFunc(str)
}
`

	formatted, err := processor.FormatImportFromSource("main.go", sourceCode)
	require.NoError(t, err)

	require.Equal(t, expectedOutput, string(formatted))
}

func TestFormatImportFromFile(t *testing.T) {
	mockResolver := NewMockResolver()

	// Add packages to the MockResolver
	mockResolver.AddPackage(&Package{
		Path: "example.com/mypkg",
		Name: "mypkg",
	})

	processor := NewProcessor(mockResolver)
	sourceFile := "main.gno"
	sourceCode := `package main

func main() {
	str := "hello, world"
	mypkg.MyFunc(str)
}`

	expectedOutput := `package main

import "example.com/mypkg"

func main() {
	str := "hello, world"
	mypkg.MyFunc(str)
}
`
	// Create a temporary directory and file
	dir := t.TempDir()
	filePath := filepath.Join(dir, sourceFile)

	err := os.WriteFile(filePath, []byte(sourceCode), 0o644)
	require.NoError(t, err)

	formatted, err := processor.FormatImportFromFile(filePath)
	require.NoError(t, err)

	require.Equal(t, expectedOutput, string(formatted))
}
