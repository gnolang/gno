package doctest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDynPackageGetter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-stdlib")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testPkgs := map[string]map[string]string{
		"std": {
			"std.gno": `
package std
type Address string
`,
		},
		"math": {
			"math.gno": `
package math
func Add(a, b int) int { return a + b }`,
			"consts.gno": `
package math
const Pi = 3.14159`,
		},
	}

	for pkgName, files := range testPkgs {
		pkgDir := filepath.Join(tempDir, pkgName)
		if err := os.Mkdir(pkgDir, 0o755); err != nil {
			t.Fatalf("failed to create package directory: %v", err)
		}
		for fileName, content := range files {
			if err := os.WriteFile(filepath.Join(pkgDir, fileName), []byte(content), 0o644); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}
		}
	}

	getter := newDynPackageLoader(tempDir)

	tests := []struct {
		name      string
		path      string
		wantPkg   bool
		wantName  string
		wantPath  string
		wantFiles int
	}{
		{
			name:      "Std package",
			path:      "std",
			wantPkg:   true,
			wantName:  "std",
			wantPath:  "std",
			wantFiles: 1,
		},
		{
			name:      "Math package",
			path:      "math",
			wantPkg:   true,
			wantName:  "math",
			wantPath:  "math",
			wantFiles: 2,
		},
		{
			name:    "Non-existent package",
			path:    "nonexistent",
			wantPkg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := getter.GetMemPackage(tt.path)
			if tt.wantPkg {
				if pkg == nil {
					t.Fatalf("expected package %s to exist", tt.path)
				}
				if pkg.Name != tt.wantName {
					t.Errorf("expected package name %s, got %s", tt.wantName, pkg.Name)
				}
				if pkg.Path != tt.wantPath {
					t.Errorf("expected package path %s, got %s", tt.wantPath, pkg.Path)
				}
				if len(pkg.Files) != tt.wantFiles {
					t.Errorf("expected %d files, got %d", tt.wantFiles, len(pkg.Files))
				}
			} else {
				if pkg != nil {
					t.Errorf("expected package %s not to exist", tt.path)
				}
			}
		})
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "Simple package",
			content: `package simple
func Foo() {}`,
			want: "simple",
		},
		{
			name: "Package with comments",
			content: `// This is a comment
package withcomments
import "fmt"
func Bar() {}`,
			want: "withcomments",
		},
		{
			name: "Package name with underscore",
			content: `package with_underscore
var x = 10`,
			want: "with_underscore",
		},
		{
			name:    "No package declaration",
			content: `func Baz() {}`,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPackageName(tt.content)
			if got != tt.want {
				t.Errorf("expected package name %s, got %s", tt.want, got)
			}
		})
	}
}
