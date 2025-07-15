package main

import (
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentPackageForCoverage(t *testing.T) {
	tests := []struct {
		name    string
		mpkg    *std.MemPackage
		wantErr bool
	}{
		{
			name: "instrument production package",
			mpkg: &std.MemPackage{
				Name: "testpkg",
				Path: "gno.land/p/demo/testpkg",
				Files: []*std.MemFile{
					{
						Name: "main.gno",
						Body: `package testpkg

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	result := 0
	for i := 0; i < b; i++ {
		result += a
	}
	return result
}`,
					},
					{
						Name: "helper.gno",
						Body: `package testpkg

func IsEven(n int) bool {
	return n%2 == 0
}`,
					},
					{
						Name: "README.md",
						Body: "# Test Package",
					},
				},
				Type: gno.MPUserProd,
			},
			wantErr: false,
		},
		{
			name: "empty package",
			mpkg: &std.MemPackage{
				Name:  "empty",
				Path:  "gno.land/p/demo/empty",
				Files: []*std.MemFile{},
				Type:  gno.MPUserProd,
			},
			wantErr: false,
		},
		{
			name: "package with only non-gno files",
			mpkg: &std.MemPackage{
				Name: "docs",
				Path: "gno.land/p/demo/docs",
				Files: []*std.MemFile{
					{
						Name: "README.md",
						Body: "# Documentation",
					},
					{
						Name: "LICENSE",
						Body: "MIT License",
					},
				},
				Type: gno.MPUserProd,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := test.NewTestOptions("", nil, nil)

			instrumentedPkg, err := instrumentPackageForCoverage(tt.mpkg, opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, instrumentedPkg)

			// Check that package metadata is preserved
			assert.Equal(t, tt.mpkg.Name, instrumentedPkg.Name)
			assert.Equal(t, tt.mpkg.Path, instrumentedPkg.Path)
			assert.Equal(t, tt.mpkg.Type, instrumentedPkg.Type)

			// Check that all files are present
			assert.Equal(t, len(tt.mpkg.Files), len(instrumentedPkg.Files))

			// Check that .gno files are instrumented and non-.gno files are unchanged
			for i, origFile := range tt.mpkg.Files {
				instrFile := instrumentedPkg.Files[i]
				assert.Equal(t, origFile.Name, instrFile.Name)

				if strings.HasSuffix(origFile.Name, ".gno") {
					// For .gno files, body should be different (instrumented)
					assert.NotEqual(t, origFile.Body, instrFile.Body)
					// Should contain coverage tracking code
					assert.Contains(t, instrFile.Body, "testing.MarkLine")
				} else {
					// For non-.gno files, body should be unchanged
					assert.Equal(t, origFile.Body, instrFile.Body)
				}
			}
		})
	}
}

func TestMergeMemPackages(t *testing.T) {
	tests := []struct {
		name     string
		prodPkg  *std.MemPackage
		testPkg  *std.MemPackage
		expected int // expected number of files
	}{
		{
			name: "merge prod and test files",
			prodPkg: &std.MemPackage{
				Name: "mypkg",
				Path: "gno.land/p/demo/mypkg",
				Files: []*std.MemFile{
					{Name: "main.gno", Body: "package mypkg\n\nfunc Main() {}"},
					{Name: "helper.gno", Body: "package mypkg\n\nfunc Helper() {}"},
				},
				Type: gno.MPUserProd,
			},
			testPkg: &std.MemPackage{
				Name: "mypkg",
				Path: "gno.land/p/demo/mypkg",
				Files: []*std.MemFile{
					{Name: "main_test.gno", Body: "package mypkg\n\nfunc TestMain() {}"},
					{Name: "helper_test.gno", Body: "package mypkg\n\nfunc TestHelper() {}"},
				},
				Type: gno.MPUserTest,
			},
			expected: 4,
		},
		{
			name: "merge with empty test package",
			prodPkg: &std.MemPackage{
				Name: "mypkg",
				Path: "gno.land/p/demo/mypkg",
				Files: []*std.MemFile{
					{Name: "main.gno", Body: "package mypkg\n\nfunc Main() {}"},
				},
				Type: gno.MPUserProd,
			},
			testPkg: &std.MemPackage{
				Name:  "mypkg",
				Path:  "gno.land/p/demo/mypkg",
				Files: []*std.MemFile{},
				Type:  gno.MPUserTest,
			},
			expected: 1,
		},
		{
			name: "merge with duplicate file names (shouldn't happen normally)",
			prodPkg: &std.MemPackage{
				Name: "mypkg",
				Path: "gno.land/p/demo/mypkg",
				Files: []*std.MemFile{
					{Name: "main.gno", Body: "package mypkg\n\nfunc Main() {}"},
					{Name: "common.gno", Body: "package mypkg\n\nvar X = 1"},
				},
				Type: gno.MPUserProd,
			},
			testPkg: &std.MemPackage{
				Name: "mypkg",
				Path: "gno.land/p/demo/mypkg",
				Files: []*std.MemFile{
					{Name: "main_test.gno", Body: "package mypkg\n\nfunc TestMain() {}"},
					{Name: "common.gno", Body: "package mypkg\n\nvar X = 2"}, // duplicate
				},
				Type: gno.MPUserTest,
			},
			expected: 3, // duplicate should be skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := mergeMemPackages(tt.prodPkg, tt.testPkg)

			require.NotNil(t, merged)
			assert.Equal(t, tt.prodPkg.Name, merged.Name)
			assert.Equal(t, tt.prodPkg.Path, merged.Path)
			assert.Equal(t, gno.MPAnyAll, merged.Type)
			assert.Equal(t, tt.expected, len(merged.Files))

			// Verify all prod files are included
			for _, prodFile := range tt.prodPkg.Files {
				found := false
				for _, mergedFile := range merged.Files {
					if mergedFile.Name == prodFile.Name {
						found = true
						assert.Equal(t, prodFile.Body, mergedFile.Body)
						break
					}
				}
				assert.True(t, found, "prod file %s not found in merged package", prodFile.Name)
			}

			// Verify test files are included (except duplicates)
			for _, testFile := range tt.testPkg.Files {
				isDuplicate := false
				for _, prodFile := range tt.prodPkg.Files {
					if prodFile.Name == testFile.Name {
						isDuplicate = true
						break
					}
				}

				if !isDuplicate {
					found := false
					for _, mergedFile := range merged.Files {
						if mergedFile.Name == testFile.Name {
							found = true
							assert.Equal(t, testFile.Body, mergedFile.Body)
							break
						}
					}
					assert.True(t, found, "test file %s not found in merged package", testFile.Name)
				}
			}
		})
	}
}

func TestCoverageIntegrationWithFilters(t *testing.T) {
	// Test that coverage instrumentation works correctly with the new filter system
	mpkg := &std.MemPackage{
		Name: "testpkg",
		Path: "gno.land/p/demo/testpkg",
		Files: []*std.MemFile{
			// Production files
			{
				Name: "math.gno",
				Body: `package testpkg

func Add(a, b int) int {
	return a + b
}`,
			},
			// Test files
			{
				Name: "math_test.gno",
				Body: `package testpkg

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("Add failed")
	}
}`,
			},
			// Integration test files
			{
				Name: "integration_test.gno",
				Body: `package testpkg_test

import (
	"testing"
	"gno.land/p/demo/testpkg"
)

func TestIntegration(t *testing.T) {
	result := testpkg.Add(10, 20)
	if result != 30 {
		t.Error("Integration test failed")
	}
}`,
			},
			// Filetest
			{
				Name: "example_filetest.gno",
				Body: `package main

import "gno.land/p/demo/testpkg"

func main() {
	println(testpkg.Add(1, 1))
}

// Output:
// 2`,
			},
		},
		Type: gno.MPUserAll,
	}

	// Test prod filter
	prodPkg := gno.MPFProd.FilterMemPackage(mpkg)
	assert.Equal(t, 1, len(prodPkg.Files))
	assert.Equal(t, "math.gno", prodPkg.Files[0].Name)
	assert.Equal(t, gno.MPUserProd, prodPkg.Type)

	// Test test filter
	testPkg := gno.MPFTest.FilterMemPackage(mpkg)
	assert.Equal(t, 2, len(testPkg.Files)) // math.gno and math_test.gno
	var hasMain, hasTest bool
	for _, f := range testPkg.Files {
		if f.Name == "math.gno" {
			hasMain = true
		}
		if f.Name == "math_test.gno" {
			hasTest = true
		}
	}
	assert.True(t, hasMain)
	assert.True(t, hasTest)
	assert.Equal(t, gno.MPUserTest, testPkg.Type)

	// Test integration filter
	integrationPkg := gno.MPFIntegration.FilterMemPackage(mpkg)
	assert.Equal(t, 1, len(integrationPkg.Files))
	assert.Equal(t, "integration_test.gno", integrationPkg.Files[0].Name)
	assert.Equal(t, gno.MPUserIntegration, integrationPkg.Type)

	// Now test coverage instrumentation with filtered packages
	opts := test.NewTestOptions("", nil, nil)

	// Instrument only production files
	instrumentedProd, err := instrumentPackageForCoverage(prodPkg, opts)
	require.NoError(t, err)
	assert.Equal(t, 1, len(instrumentedProd.Files))
	assert.Contains(t, instrumentedProd.Files[0].Body, "testing.MarkLine")

	// Merge instrumented prod with test files
	merged := mergeMemPackages(instrumentedProd, testPkg)
	assert.Equal(t, 2, len(merged.Files))

	// Verify instrumented prod file is in merged package
	var hasInstrumentedProd bool
	for _, f := range merged.Files {
		if f.Name == "math.gno" && strings.Contains(f.Body, "testing.MarkLine") {
			hasInstrumentedProd = true
		}
	}
	assert.True(t, hasInstrumentedProd)
}
