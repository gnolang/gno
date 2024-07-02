package gnolang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMemPackage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testpkg")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create valid files
	validFiles := []string{"file1.gno", "README.md", "LICENSE", "gno.mod"}
	for _, f := range validFiles {
		err := os.WriteFile(filepath.Join(tempDir, f), []byte(`
		package main
		
		import (
			"gno.land/p/demo/ufmt"
		)
		
		func main() {
			ufmt.Printfln("Hello, World!")
		}`), 0o644)
		require.NoError(t, err)
	}

	// Create invalid files
	invalidFiles := []string{".hiddenfile", "unsupported.txt"}
	for _, f := range invalidFiles {
		err := os.WriteFile(filepath.Join(tempDir, f), []byte("content"), 0o644)
		require.NoError(t, err)
	}

	// Test Case 1: Valid Package Directory
	memPkg := ReadMemPackage(tempDir, "testpkg")
	require.NotNil(t, memPkg)
	assert.Len(t, memPkg.Files, len(validFiles), "MemPackage should contain only valid files")

	// Test Case 2: Non-existent Directory
	assert.Panics(t, func() {
		ReadMemPackage("/non/existent/dir", "testpkg")
	}, "Expected panic for non-existent directory")
}
