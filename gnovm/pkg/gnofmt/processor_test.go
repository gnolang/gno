// For convenient purposes, more tests have been created using `testscripts` format and are located in `gnovm/cmd/testdata/gno_fmt/` folder

package gnofmt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
func TestTrimTrailingWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{
			name:  "no trailing whitespace",
			input: []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
			want:  []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
		},
		{
			name:  "simple trailing spaces",
			input: []byte("package main  \n\nfunc main() { \n\tprintln(\"hello\")\t\n} "),
			want:  []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
		},
		{
			name:  "trailing tabs",
			input: []byte("package main\t\n\nfunc main() {\t\n\tprintln(\"hello\")\t\n}\t"),
			want:  []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
		},
		{
			name:  "mixed spaces and tabs",
			input: []byte("package main \t \n\nfunc main() { \t\n\tprintln(\"hello\")\t \n} \t "),
			want:  []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
		},
		{
			name:  "empty lines with whitespace",
			input: []byte("package main\n \n\t\nfunc main() {\n\tprintln(\"hello\")\n}"),
			want:  []byte("package main\n\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
		},
		{
			name:  "no final newline",
			input: []byte("package main   "),
			want:  []byte("package main"),
		},
		{
			name:  "multiple trailing newlines",
			input: []byte("package main\n\n\n"),
			want:  []byte("package main\n\n\n"),
		},
		{
			name:  "single line with trailing whitespace",
			input: []byte("single line   "),
			want:  []byte("single line"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimTrailingWhitespace(tt.input)
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
