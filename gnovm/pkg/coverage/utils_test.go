package coverage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTestFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pkgPath string
		want    bool
	}{
		{"file1_test.gno", true},
		{"file1_testing.gno", true},
		{"file1.gno", false},
		{"random_test.go", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.pkgPath, func(t *testing.T) {
			t.Parallel()
			got := IsTestFile(tt.pkgPath)
			if got != tt.want {
				t.Errorf("isTestFile(%s) = %v, want %v", tt.pkgPath, got, tt.want)
			}
		})
	}
}

func TestFindAbsoluteFilePath(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()

	examplesDir := filepath.Join(rootDir, "examples")
	stdlibsDir := filepath.Join(rootDir, "gnovm", "stdlibs")

	if err := os.MkdirAll(examplesDir, 0o755); err != nil {
		t.Fatalf("failed to create examples directory: %v", err)
	}
	if err := os.MkdirAll(stdlibsDir, 0o755); err != nil {
		t.Fatalf("failed to create stdlibs directory: %v", err)
	}

	exampleFile := filepath.Join(examplesDir, "example.gno")
	stdlibFile := filepath.Join(stdlibsDir, "stdlib.gno")
	if _, err := os.Create(exampleFile); err != nil {
		t.Fatalf("failed to create example file: %v", err)
	}
	if _, err := os.Create(stdlibFile); err != nil {
		t.Fatalf("failed to create stdlib file: %v", err)
	}

	c := New(rootDir)

	tests := []struct {
		name         string
		filePath     string
		expectedPath string
		expectError  bool
	}{
		{
			name:         "File in examples directory",
			filePath:     "example.gno",
			expectedPath: exampleFile,
			expectError:  false,
		},
		{
			name:         "File in stdlibs directory",
			filePath:     "stdlib.gno",
			expectedPath: stdlibFile,
			expectError:  false,
		},
		{
			name:         "Non-existent file",
			filePath:     "nonexistent.gno",
			expectedPath: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actualPath, err := findAbsFilePath(c, tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
				if actualPath != tt.expectedPath {
					t.Errorf("expected path %s, but got %s", tt.expectedPath, actualPath)
				}
			}
		})
	}
}

func TestFindAbsoluteFilePathCache(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "example.gno")
	if err := os.WriteFile(testFilePath, []byte("test content"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	covData := New(tempDir)

	// 1st run: search from file system
	path1, err := findAbsFilePath(covData, "example.gno")
	if err != nil {
		t.Fatalf("failed to find absolute file path: %v", err)
	}
	assert.Equal(t, testFilePath, path1)

	// 2nd run: use cache
	path2, err := findAbsFilePath(covData, "example.gno")
	if err != nil {
		t.Fatalf("failed to find absolute file path: %v", err)
	}

	assert.Equal(t, testFilePath, path2)
	if len(covData.pathCache) != 1 {
		t.Fatalf("expected 1 path in cache, got %d", len(covData.pathCache))
	}
}
