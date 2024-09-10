package gnolang

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddHit(t *testing.T) {
	tests := []struct {
		name         string
		initialData  *CoverageData
		pkgPath      string
		line         int
		expectedHits int
	}{
		{
			name:         "Add hit to new file",
			initialData:  NewCoverageData(""),
			pkgPath:      "file1.gno",
			line:         10,
			expectedHits: 1,
		},
		{
			name: "Add hit to existing file and line",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {HitLines: map[int]int{10: 1}},
				},
			},
			pkgPath:      "file1.gno",
			line:         10,
			expectedHits: 2,
		},
		{
			name: "Add hit to new line in existing file",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {HitLines: map[int]int{10: 1}},
				},
			},
			pkgPath:      "file1.gno",
			line:         20,
			expectedHits: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialData.AddHit(tt.pkgPath, tt.line)
			fileCoverage := tt.initialData.Files[tt.pkgPath]

			// Validate the hit count for the specific line
			if fileCoverage.HitLines[tt.line] != tt.expectedHits {
				t.Errorf("got %d hits for line %d, want %d", fileCoverage.HitLines[tt.line], tt.line, tt.expectedHits)
			}
		})
	}
}

func TestAddFile(t *testing.T) {
	tests := []struct {
		name          string
		pkgPath       string
		totalLines    int
		initialData   *CoverageData
		expectedTotal int
	}{
		{
			name:          "Add new file",
			pkgPath:       "file1.gno",
			totalLines:    100,
			initialData:   NewCoverageData(""),
			expectedTotal: 100,
		},
		{
			name:          "Do not add test file *_test.gno",
			pkgPath:       "file1_test.gno",
			totalLines:    100,
			initialData:   NewCoverageData(""),
			expectedTotal: 0,
		},
		{
			name:          "Do not add test file *_testing.gno",
			pkgPath:       "file1_testing.gno",
			totalLines:    100,
			initialData:   NewCoverageData(""),
			expectedTotal: 0,
		},
		{
			name:       "Update existing file's total lines",
			pkgPath:    "file1.gno",
			totalLines: 200,
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {TotalLines: 100, HitLines: map[int]int{10: 1}},
				},
			},
			expectedTotal: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialData.AddFile(tt.pkgPath, tt.totalLines)
			if tt.pkgPath == "file1_test.gno" && len(tt.initialData.Files) != 0 {
				t.Errorf("expected no files to be added for test files")
			} else {
				if fileCoverage, ok := tt.initialData.Files[tt.pkgPath]; ok {
					if fileCoverage.TotalLines != tt.expectedTotal {
						t.Errorf("got %d total lines, want %d", fileCoverage.TotalLines, tt.expectedTotal)
					}
				} else if len(tt.initialData.Files) > 0 {
					t.Errorf("expected file not added")
				}
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
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
		t.Run(tt.pkgPath, func(t *testing.T) {
			got := isTestFile(tt.pkgPath)
			if got != tt.want {
				t.Errorf("isTestFile(%s) = %v, want %v", tt.pkgPath, got, tt.want)
			}
		})
	}
}

func TestReport(t *testing.T) {
	tests := []struct {
		name           string
		initialData    *CoverageData
		expectedOutput string
	}{
		{
			name: "Print results with one file",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {TotalLines: 100, HitLines: map[int]int{10: 1, 20: 1}},
				},
			},
			expectedOutput: "Coverage Results:\nfile1.gno: 2.00% (2/100 lines)\n",
		},
		{
			name: "Print results with multiple files",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {TotalLines: 100, HitLines: map[int]int{10: 1, 20: 1}},
					"file2.gno": {TotalLines: 200, HitLines: map[int]int{30: 1}},
				},
			},
			expectedOutput: "Coverage Results:\nfile1.gno: 2.00% (2/100 lines)\nfile2.gno: 0.50% (1/200 lines)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origStdout := os.Stdout

			r, w, _ := os.Pipe()
			os.Stdout = w

			tt.initialData.Report()

			w.Close()
			os.Stdout = origStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			got := buf.String()
			if got != tt.expectedOutput {
				t.Errorf("got %q, want %q", got, tt.expectedOutput)
			}
		})
	}
}

type mockNode struct {
	line   int
	column int
}

func (m *mockNode) assertNode()                               {}
func (m *mockNode) String() string                            { return "" }
func (m *mockNode) Copy() Node                                { return &mockNode{} }
func (m *mockNode) GetLabel() Name                            { return "mockNode" }
func (m *mockNode) SetLabel(n Name)                           {}
func (m *mockNode) HasAttribute(n interface{}) bool           { return false }
func (m *mockNode) GetAttribute(n interface{}) interface{}    { return nil }
func (m *mockNode) SetAttribute(n interface{}, v interface{}) {}
func (m *mockNode) GetLine() int                              { return m.line }
func (m *mockNode) SetLine(l int)                             {}
func (m *mockNode) GetColumn() int                            { return m.column }
func (m *mockNode) SetColumn(c int)                           {}

var _ Node = &mockNode{}

func TestRecordCoverage(t *testing.T) {
	tests := []struct {
		name           string
		node           Node
		currentPackage string
		currentFile    string
		expectedLoc    Location
		expectedHits   map[string]map[int]int
	}{
		{
			name:           "Basic node coverage",
			node:           &mockNode{line: 10, column: 5},
			currentPackage: "testpkg",
			currentFile:    "testfile.gno",
			expectedLoc:    Location{PkgPath: "testpkg", File: "testfile.gno", Line: 10, Column: 5},
			expectedHits:   map[string]map[int]int{"testpkg/testfile.gno": {10: 1}},
		},
		{
			name:           "Nil node",
			node:           nil,
			currentPackage: "testpkg",
			currentFile:    "testfile.gno",
			expectedLoc:    Location{},
			expectedHits:   map[string]map[int]int{},
		},
		{
			name:           "Multiple hits on same line",
			node:           &mockNode{line: 15, column: 3},
			currentPackage: "testpkg",
			currentFile:    "testfile.gno",
			expectedLoc:    Location{PkgPath: "testpkg", File: "testfile.gno", Line: 15, Column: 3},
			expectedHits:   map[string]map[int]int{"testpkg/testfile.gno": {15: 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Machine{
				CurrentPackage: tt.currentPackage,
				CurrentFile:    tt.currentFile,
				Coverage:       NewCoverageData(""),
			}

			// First call to set up initial state for "Multiple hits on same line" test
			if tt.name == "Multiple hits on same line" {
				m.recordCoverage(tt.node)
			}

			loc := m.recordCoverage(tt.node)

			assert.Equal(t, tt.expectedLoc, loc, "Location should match")

			for file, lines := range tt.expectedHits {
				for line, hits := range lines {
					actualHits, exists := m.Coverage.Files[file].HitLines[line]
					assert.True(t, exists, "Line should be recorded in coverage data")
					assert.Equal(t, hits, actualHits, "Number of hits should match")
				}
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name         string
		coverageData *CoverageData
		expectedJSON string
	}{
		{
			name: "Single file with hits",
			coverageData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {
						TotalLines: 100,
						HitLines:   map[int]int{10: 1, 20: 2},
					},
				},
			},
			expectedJSON: `{
  "files": {
    "file1.gno": {
      "total_lines": 100,
      "hit_lines": {
        "10": 1,
        "20": 2
      }
    }
  }
}`,
		},
		{
			name: "Multiple files with hits",
			coverageData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {
						TotalLines: 100,
						HitLines:   map[int]int{10: 1, 20: 2},
					},
					"file2.gno": {
						TotalLines: 200,
						HitLines:   map[int]int{30: 3},
					},
				},
			},
			expectedJSON: `{
  "files": {
    "file1.gno": {
      "total_lines": 100,
      "hit_lines": {
        "10": 1,
        "20": 2
      }
    },
    "file2.gno": {
      "total_lines": 200,
      "hit_lines": {
        "30": 3
      }
    }
  }
}`,
		},
		{
			name: "No files",
			coverageData: &CoverageData{
				Files: map[string]FileCoverage{},
			},
			expectedJSON: `{
  "files": {}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := tt.coverageData.ToJSON()
			assert.NoError(t, err)

			var got map[string]interface{}
			var expected map[string]interface{}

			err = json.Unmarshal(jsonData, &got)
			assert.NoError(t, err)

			err = json.Unmarshal([]byte(tt.expectedJSON), &expected)
			assert.NoError(t, err)

			assert.Equal(t, expected, got)
		})
	}
}

func TestDetermineRealPath(t *testing.T) {
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

	c := &CoverageData{
		RootDir: rootDir,
	}

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
		t.Run(tt.name, func(t *testing.T) {
			actualPath, err := c.determineRealPath(tt.filePath)

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

func TestDetectExecutableLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[int]bool
		wantErr bool
	}{
		{
			name: "Simple function",
			content: `
package main

func main() {
	x := 5
	if x > 3 {
		println("Greater")
	}
}`,
			want: map[int]bool{
				5: true, // x := 5
				6: true, // if x > 3
				7: true, // println("Greater")
			},
			wantErr: false,
		},
		{
			name: "Function with loop",
			content: `
package main

func loopFunction() {
	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			continue
		}
		println(i)
	}
}`,
			want: map[int]bool{
				5: true, // for i := 0; i < 5; i++
				6: true, // if i%2 == 0
				7: true, // continue
				9: true, // println(i)
			},
			wantErr: false,
		},
		{
			name: "Only declarations",
			content: `
package main

import "fmt"

var x int

type MyStruct struct {
	field int
}`,
			want:    map[int]bool{},
			wantErr: false,
		},
		{
			name: "Invalid Go code",
			content: `
This is not valid Go code
It should result in an error`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectExecutableLines(tt.content)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
