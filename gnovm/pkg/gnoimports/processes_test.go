// For convenient purposes, more tests have been created using `testscripts` format and are located in `gnovm/cmd/testdata/gno_fmt/` folder

package gnoimports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatImportFromSource(t *testing.T) {
	t.Parallel()

	mockResolver := newMockResolver()

	mp := newMockedPackage("example.com/mypkg", "mypkg")
	pkgcontent := `package mypkg

func MyFunc(str string) string{
	return "Hello: "+str
}`
	mp.AddFile("my.gno", []byte(pkgcontent))
	mockResolver.AddPackage(mp)

	sourceCode := `package main

func main() {
	str := "hello, world"
	mypkg.MyFunc(str)
}`

	// Add packages to the MockResolver
	processor := NewProcessor(mockResolver)
	formatted, err := processor.FormatImportFromSource("main.go", sourceCode)
	require.NoError(t, err)

	expectedOutput := `package main

import "example.com/mypkg"

func main() {
	str := "hello, world"
	mypkg.MyFunc(str)
}
`

	require.Equal(t, expectedOutput, string(formatted))
}

func TestFormatImportFromFile(t *testing.T) {
	t.Parallel()

	mockResolver := newMockResolver()

	// Add packages to the MockResolver
	mp := newMockedPackage("example.com/mypkg", "mypkg")
	pkgcontent := `package mypkg

func MyFunc(str string) string{
	return "Hello: "+str
}`
	mp.AddFile("my.gno", []byte(pkgcontent))
	mockResolver.AddPackage(mp)

	processor := NewProcessor(mockResolver)
	sourceFile := "main.gno"
	sourceCode := `package main

func main() {
	str := "hello, world"
	println(mypkg.MyFunc(str))
}`

	expectedOutput := `package main

import "example.com/mypkg"

func main() {
	str := "hello, world"
	println(mypkg.MyFunc(str))
}
`
	// Create a temporary directory and file
	dir := t.TempDir()
	filePath := filepath.Join(dir, sourceFile)

	err := os.WriteFile(filePath, []byte(sourceCode), 0o644)
	require.NoError(t, err)

	formatted, err := processor.FormatFile(filePath)
	require.NoError(t, err)

	require.Equal(t, expectedOutput, string(formatted))
}
