// For convenient purposes, more tests have been created using `testscripts` format and are located in `gnovm/cmd/testdata/gno_fmt/` folder

package gnofmt

import (
	"go/token"
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

func TestCheckPackageConsistency(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	t.Run("consistent", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a.gno"), []byte("package foo\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "b.gno"), []byte("package foo\n"), 0o644))

		ok, err := CheckPackageConsistency(fset, dir)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("inconsistent", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a.gno"), []byte("package foo\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "b.gno"), []byte("package bar\n"), 0o644))

		ok, err := CheckPackageConsistency(fset, dir)
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("single_file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a.gno"), []byte("package foo\n"), 0o644))

		ok, err := CheckPackageConsistency(fset, dir)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("test_files_ignored", func(t *testing.T) {
		t.Parallel()
		// Test files (_test.gno) should not affect consistency.
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a.gno"), []byte("package foo\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a_test.gno"), []byte("package bar\n"), 0o644))

		ok, err := CheckPackageConsistency(fset, dir)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("empty_dir", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		ok, err := CheckPackageConsistency(fset, dir)
		require.NoError(t, err)
		require.True(t, ok)
	})

	// Verify the real gnovm/tests/files/ directory is detected as inconsistent.
	// This directory contains independent filetests with different package names
	// (e.g. "main" and "test"), which is the primary real-world use case.
	t.Run("real_tests_files", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join("..", "..", "tests", "files")
		// Ensure the directory exists so this test isn't silently skipped.
		_, err := os.Stat(dir)
		require.NoError(t, err, "gnovm/tests/files/ directory should exist")

		ok, err := CheckPackageConsistency(fset, dir)
		require.NoError(t, err)
		require.False(t, ok, "gnovm/tests/files/ should have mixed package names")
	})
}

func TestFormatFileConflictingPackages(t *testing.T) {
	t.Parallel()

	mockResolver := newMockResolver()
	processor := NewProcessor(mockResolver)

	// Create a temp dir with two .gno files having different package names,
	// simulating a filetest directory like gnovm/tests/files/.
	dir := t.TempDir()

	// Deliberately unformatted: bad spacing, missing indentation, no spaces around operators.
	file1 := `package main

func    main(   ) {
x:=1
  y   :=2
if x==y{
println(   "equal")
} else    {
println("not equal"  )
}
}
`
	file1Expected := `package main

func main() {
	x := 1
	y := 2
	if x == y {
		println("equal")
	} else {
		println("not equal")
	}
}
`

	file2 := `package other

func   Foo()  string {
return    "foo"
}
`
	file2Expected := `package other

func Foo() string {
	return "foo"
}
`
	err := os.WriteFile(filepath.Join(dir, "file1.gno"), []byte(file1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "file2.gno"), []byte(file2), 0o644)
	require.NoError(t, err)

	// FormatFile should succeed despite conflicting package names,
	// falling back to per-file formatting, and actually format the code.
	formatted, err := processor.FormatFile(filepath.Join(dir, "file1.gno"))
	require.NoError(t, err)
	require.Equal(t, file1Expected, string(formatted))

	formatted, err = processor.FormatFile(filepath.Join(dir, "file2.gno"))
	require.NoError(t, err)
	require.Equal(t, file2Expected, string(formatted))
}
