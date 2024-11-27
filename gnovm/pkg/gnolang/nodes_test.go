package gnolang_test

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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
	t.Parallel()
	tests := []struct {
		name        string
		files       map[string]string // map[filename]content
		pkgPath     string
		shouldPanic bool
		wantPkgName string
	}{
		{
			name: "valid package - math",
			files: map[string]string{
				"math.gno": `package math
					func Add(a, b int) int { return a + b }`,
			},
			pkgPath:     "std/math",
			shouldPanic: false,
			wantPkgName: "math",
		},
		{
			name: "valid package - bytealg",
			files: map[string]string{
				"bytealg.gno": `package bytealg
					func Compare(a, b []byte) int { return 0 }`,
			},
			pkgPath:     "std/bytealg",
			shouldPanic: false,
			wantPkgName: "bytealg",
		},
		{
			name: "valid package - sha256",
			files: map[string]string{
				"sha256.gno": `package sha256
					func Sum256(data []byte) [32]byte { var sum [32]byte; return sum }`,
			},
			pkgPath:     "crypto/sha256",
			shouldPanic: false,
			wantPkgName: "sha256",
		},
		{
			name: "nested package - foo/bar",
			files: map[string]string{
				"bar.gno": `package bar
					func DoSomething() {}`,
			},
			pkgPath:     "gno.land/foo/bar",
			shouldPanic: false,
			wantPkgName: "bar",
		},
		{
			name: "package with README and LICENSE",
			files: map[string]string{
				"math.gno": `package math
					func Add(a, b int) int { return a + b }`,
				"README.md": "# Math Package",
				"LICENSE":   "MIT License",
			},
			pkgPath:     "std/math",
			shouldPanic: false,
			wantPkgName: "math",
		},
		{
			name: "stdlib with .go files",
			files: map[string]string{
				"math.gno": `package math
					func Add(a, b int) int { return a + b }`,
				"native.go": `package math
					func NativeAdd(a, b int) int { return a + b }`,
			},
			pkgPath:     "std/math",
			shouldPanic: false,
			wantPkgName: "math",
		},
		{
			name: "stdlib with rejected .gen.go files",
			files: map[string]string{
				"math.gno": `package math
					func Add(a, b int) int { return a + b }`,
				"generated.gen.go": `package math
					func GeneratedFunc() {}`,
			},
			pkgPath:     "std/math",
			shouldPanic: false,
			wantPkgName: "math",
		},
		{
			name: "valid nested package - gnoland/xxx/foo/bar",
			files: map[string]string{
				"bar.gno": `package bar
					func DoSomething() string { return "hello" }`,
			},
			pkgPath:     "gno.land/xxx/foo/bar",
			shouldPanic: false,
			wantPkgName: "bar",
		},
		{
			name: "valid package - gno.land/p/demo/tests",
			files: map[string]string{
				"tests.gno": `package tests
					const World = "world"`,
			},
			pkgPath:     "gno.land/p/demo/tests",
			shouldPanic: false,
			wantPkgName: "tests",
		},
		{
			name: "foo/bar with empty foo directory",
			files: map[string]string{
				"bar.gno": `package bar
					func DoSomething() string { return "hello" }`,
			},
			pkgPath:     "gno.land/r/foo/bar",
			shouldPanic: false,
			wantPkgName: "bar",
		},
		{
			name: "invalid package - internal",
			files: map[string]string{
				"internal.gno": `package internal
					func someFunc() {}`,
			},
			pkgPath:     "std/internal",
			shouldPanic: true,
		},
		{
			name: "invalid package - crypto",
			files: map[string]string{
				"crypto.gno": `package crypto
					func Hash(data []byte) []byte { return nil }`,
			},
			pkgPath:     "std/crypto",
			shouldPanic: true,
		},
		{
			name: "empty package without gno files",
			files: map[string]string{
				"README.md": "# Empty Package",
				"LICENSE": "MIT License",
			},
			pkgPath:     "gno.land/r/foo",
			shouldPanic: true,
			wantPkgName: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir, err := os.MkdirTemp("", "test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			for fname, content := range tt.files {
				fpath := filepath.Join(tmpDir, fname)
				dir := filepath.Dir(fpath)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for package %s", tt.pkgPath)
					}
				}()
			}

			memPkg := gnolang.ReadMemPackage(tmpDir, tt.pkgPath)

			if !tt.shouldPanic {
				if memPkg == nil {
					t.Fatal("expected non-nil MemPackage")
				}

				if memPkg.Name != tt.wantPkgName {
					t.Errorf("got package name %q, want %q", memPkg.Name, tt.wantPkgName)
				}

				if memPkg.Path != tt.pkgPath {
					t.Errorf("got package path %q, want %q", memPkg.Path, tt.pkgPath)
				}

				expectedFiles := 0
				for fname := range tt.files {
					if strings.HasSuffix(fname, ".gen.go") {
						continue
					}
					if strings.HasSuffix(fname, ".gno") ||
						fname == "README.md" ||
						fname == "LICENSE" ||
						(gnolang.IsStdlib(tt.pkgPath) && strings.HasSuffix(fname, ".go")) {
						expectedFiles++
					}
				}
				if len(memPkg.Files) != expectedFiles {
					t.Errorf("got %d files, want %d", len(memPkg.Files), expectedFiles)
				}
			}
		})
	}
}

func TestInjectNativeMethod(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name        string
        files       map[string]string
        pkgPath     string
        shouldPanic bool
    }{
        {
            name: "inject to valid package",
            files: map[string]string{
                "foo.gno": `package foo
                    func RegularFunc() string { return "hello" }`,
            },
            pkgPath:     "gno.land/r/test/foo",
            shouldPanic: false,
        },
        {
            name: "inject to empty package",
            files: map[string]string{
                "README.md": "# Empty Package",
            },
            pkgPath:     "gno.land/r/test/empty",
            shouldPanic: true,
        },
        {
            name: "inject to foo in foo/bar structure",
            files: map[string]string{
                "bar/bar.gno": `package bar
                    func BarFunc() string { return "bar" }`,
            },
            pkgPath:     "gno.land/r/test/foo",
            shouldPanic: true,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            
            tmpDir, err := os.MkdirTemp("", "test-*")
            if err != nil {
                t.Fatal(err)
            }
            defer os.RemoveAll(tmpDir)

            for fname, content := range tt.files {
                fpath := filepath.Join(tmpDir, fname)
                dir := filepath.Dir(fpath)
                if err := os.MkdirAll(dir, 0o755); err != nil {
                    t.Fatal(err)
                }
                if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
                    t.Fatal(err)
                }
            }

            defer func() {
                if r := recover(); r != nil {
                    if !tt.shouldPanic {
                        t.Errorf("unexpected panic: %v", r)
                    }
                } else if tt.shouldPanic {
                    t.Error("expected panic, but got none")
                }
            }()

            memPkg := gnolang.ReadMemPackage(tmpDir, tt.pkgPath)
            fset := gnolang.ParseMemPackage(memPkg)
            pkgNode := gnolang.NewPackageNode(gnolang.Name(memPkg.Name), tt.pkgPath, fset)

            pkgNode.DefineNative("NativeMethod",
                gnolang.FieldTypeExprs{},
                gnolang.FieldTypeExprs{},
                func(m *gnolang.Machine) {},
            )
        })
    }
}
