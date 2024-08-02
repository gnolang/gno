import (
	"math"
  "os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"

  "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticBlock_Define2_MaxNames(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			panicString, ok := r.(string)
			if !ok {
				t.Errorf("expected panic string, got %v", r)
			}

			if panicString != "too many variables in block" {
				t.Errorf("expected panic string to be 'too many variables in block', got '%s'", panicString)
			}

			return
		}

		// If it didn't panic, fail.
		t.Errorf("expected panic when exceeding maximum number of names")
	}()

	staticBlock := new(gnolang.StaticBlock)
	staticBlock.NumNames = math.MaxUint16 - 1
	staticBlock.Names = make([]gnolang.Name, staticBlock.NumNames)

	// Adding one more is okay.
	staticBlock.Define2(false, gnolang.Name("a"), gnolang.BoolType, gnolang.TypedValue{T: gnolang.BoolType})
	if staticBlock.NumNames != math.MaxUint16 {
		t.Errorf("expected NumNames to be %d, got %d", math.MaxUint16, staticBlock.NumNames)
	}
	if len(staticBlock.Names) != math.MaxUint16 {
		t.Errorf("expected len(Names) to be %d, got %d", math.MaxUint16, len(staticBlock.Names))
	}

	// This one should panic because the maximum number of names has been reached.
	staticBlock.Define2(false, gnolang.Name("a"), gnolang.BoolType, gnolang.TypedValue{T: gnolang.BoolType})
}

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