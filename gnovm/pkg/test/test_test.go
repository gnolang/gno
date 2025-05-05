package test

import (
	"testing"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
)

func TestLoadTestFuncs(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		fileBody string
		want     []testFunc
	}{
		{
			name:    "empty file set",
			pkgName: "test",
			fileBody: `
				package test
			`,
			want: nil,
		},
		{
			name:    "single test function",
			pkgName: "test",
			fileBody: `
				package test
				func TestSomething(t *testing.T) {}
			`,
			want: []testFunc{
				{Package: "test", Name: "TestSomething"},
			},
		},
		{
			name:    "multiple test functions",
			pkgName: "test",
			fileBody: `
				package test
				func TestOne(t *testing.T) {}
				func TestTwo(t *testing.T) {}
				func helper() {}
				func TestThree(t *testing.T) {}
			`,
			want: []testFunc{
				{Package: "test", Name: "TestOne"},
				{Package: "test", Name: "TestTwo"},
				{Package: "test", Name: "TestThree"},
			},
		},
		{
			name:    "non-test functions",
			pkgName: "test",
			fileBody: `
				package test
				func helper() {}
				func regular() {}
				func test() {}
			`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test file
			file, err := gno.ParseFile("test.gno", tt.fileBody)
			if err != nil {
				t.Fatalf("failed to parse test file: %v", err)
			}

			fileSet := &gno.FileSet{}
			fileSet.AddFiles(file)

			// Run the function
			got := loadTestFuncs(tt.pkgName, fileSet)

			// Assert the results
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseMemPackageTests(t *testing.T) {
	tests := []struct {
		name        string
		memPkg      *gnovm.MemPackage
		wantTSet    bool // Whether there are regular test files
		wantITSet   bool // Whether there are integration test files
		wantITFiles int  // Number of integration test files
		wantFTFiles int  // Number of file test files
	}{
		{
			name: "empty package",
			memPkg: &gnovm.MemPackage{
				Name:  "test",
				Path:  "test",
				Files: nil,
			},
			wantTSet:    false,
			wantITSet:   false,
			wantITFiles: 0,
			wantFTFiles: 0,
		},
		{
			name: "case with only regular test files",
			memPkg: &gnovm.MemPackage{
				Name: "test",
				Path: "test",
				Files: []*gnovm.MemFile{
					{
						Name: "something_test.gno",
						Body: `package test
							func TestSomething(t *testing.T) {}`,
					},
				},
			},
			wantTSet:    true,
			wantITSet:   false,
			wantITFiles: 0,
			wantFTFiles: 0,
		},
		{
			name: "case with only integration test files",
			memPkg: &gnovm.MemPackage{
				Name: "test",
				Path: "test",
				Files: []*gnovm.MemFile{
					{
						Name: "integration_test.gno",
						Body: `package test_test
							func TestIntegration(t *testing.T) {}`,
					},
				},
			},
			wantTSet:    false,
			wantITSet:   true,
			wantITFiles: 1,
			wantFTFiles: 0,
		},
		{
			name: "case with only file tests",
			memPkg: &gnovm.MemPackage{
				Name: "test",
				Path: "test",
				Files: []*gnovm.MemFile{
					{
						Name: "something_filetest.gno",
						Body: `package test
							// File test content`,
					},
				},
			},
			wantTSet:    false,
			wantITSet:   false,
			wantITFiles: 0,
			wantFTFiles: 1,
		},
		{
			name: "case with all types of test files",
			memPkg: &gnovm.MemPackage{
				Name: "test",
				Path: "test",
				Files: []*gnovm.MemFile{
					{
						Name: "normal_test.gno",
						Body: `package test
							func TestNormal(t *testing.T) {}`,
					},
					{
						Name: "integration_test.gno",
						Body: `package test_test
							func TestIntegration(t *testing.T) {}`,
					},
					{
						Name: "file_filetest.gno",
						Body: `package test
							// File test content`,
					},
					{
						Name: "regular.gno",
						Body: `package test
							func Regular() {}`,
					},
				},
			},
			wantTSet:    true,
			wantITSet:   true,
			wantITFiles: 1,
			wantFTFiles: 1,
		},
		{
			name: "ignore files with incorrect extensions",
			memPkg: &gnovm.MemPackage{
				Name: "test",
				Path: "test",
				Files: []*gnovm.MemFile{
					{
						Name: "test.txt",
						Body: "Any content",
					},
				},
			},
			wantTSet:    false,
			wantITSet:   false,
			wantITFiles: 0,
			wantFTFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tset, itset, itfiles, ftfiles := parseMemPackageTests(tt.memPkg)

			// Check regular test files
			if tt.wantTSet {
				assert.NotNil(t, tset)
				assert.Greater(t, len(tset.Files), 0)
			} else {
				assert.Equal(t, 0, len(tset.Files))
			}

			// Check integration test files
			if tt.wantITSet {
				assert.NotNil(t, itset)
				assert.Greater(t, len(itset.Files), 0)
			} else {
				assert.Equal(t, 0, len(itset.Files))
			}

			// Check number of integration test files
			assert.Equal(t, tt.wantITFiles, len(itfiles))

			// Check number of file test files
			assert.Equal(t, tt.wantFTFiles, len(ftfiles))
		})
	}
}
