package gnolang

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
)

func TestCoverageDataUpdateHit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		initialData     *CoverageData
		pkgPath         string
		line            int
		expectedHits    int
		executableLines map[int]bool
	}{
		{
			name: "Add hit to existing file and executable line",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {
						HitLines:        map[int]int{10: 1},
						ExecutableLines: map[int]bool{10: true, 20: true},
					},
				},
			},
			pkgPath:         "file1.gno",
			line:            10,
			expectedHits:    2,
			executableLines: map[int]bool{10: true, 20: true},
		},
		{
			name: "Add hit to new executable line in existing file",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {
						HitLines:        map[int]int{10: 1},
						ExecutableLines: map[int]bool{10: true, 20: true},
					},
				},
			},
			pkgPath:         "file1.gno",
			line:            20,
			expectedHits:    1,
			executableLines: map[int]bool{10: true, 20: true},
		},
		{
			name: "Add hit to non-executable line",
			initialData: &CoverageData{
				Files: map[string]FileCoverage{
					"file1.gno": {
						HitLines:        map[int]int{10: 1},
						ExecutableLines: map[int]bool{10: true},
					},
				},
			},
			pkgPath:         "file1.gno",
			line:            20,
			expectedHits:    0,
			executableLines: map[int]bool{10: true},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Set executable lines
			fileCoverage := tt.initialData.Files[tt.pkgPath]
			fileCoverage.ExecutableLines = tt.executableLines
			tt.initialData.Files[tt.pkgPath] = fileCoverage

			tt.initialData.updateHit(tt.pkgPath, tt.line)
			updatedFileCoverage := tt.initialData.Files[tt.pkgPath]

			// Validate the hit count for the specific line
			actualHits := updatedFileCoverage.HitLines[tt.line]
			if actualHits != tt.expectedHits {
				t.Errorf("got %d hits for line %d, want %d", actualHits, tt.line, tt.expectedHits)
			}

			// Check if non-executable lines are not added to HitLines
			if !tt.executableLines[tt.line] && actualHits > 0 {
				t.Errorf("non-executable line %d was added to HitLines", tt.line)
			}
		})
	}
}

func TestAddFile(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.initialData.addFile(tt.pkgPath, tt.totalLines)
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
			got := isTestFile(tt.pkgPath)
			if got != tt.want {
				t.Errorf("isTestFile(%s) = %v, want %v", tt.pkgPath, got, tt.want)
			}
		})
	}
}

type nopCloser struct {
	*bytes.Buffer
}

func (nopCloser) Close() error { return nil }

func TestCoverageData_GenerateReport(t *testing.T) {
	coverageData := &CoverageData{
		Files: map[string]FileCoverage{
			"c.gno": {TotalLines: 100, HitLines: map[int]int{1: 1, 2: 1}},
			"a.gno": {TotalLines: 50, HitLines: map[int]int{1: 1}},
			"b.gno": {TotalLines: 75, HitLines: map[int]int{1: 1, 2: 1, 3: 1}},
		},
	}

	var buf bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(nopCloser{Buffer: &buf})

	coverageData.Report(io)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// check if the output is sorted
	assert.Equal(t, 3, len(lines))
	assert.Contains(t, lines[0], "a.gno")
	assert.Contains(t, lines[1], "b.gno")
	assert.Contains(t, lines[2], "c.gno")

	// check if the format is correct
	for _, line := range lines {
		assert.Regexp(t, `^\x1b\[\d+m\d+\.\d+% \[\s*\d+/\d+\] .+\.gno\x1b\[0m$`, line)
	}

	// check if the coverage percentage is correct
	assert.Contains(t, lines[0], "2.0% [   1/50] a.gno")
	assert.Contains(t, lines[1], "4.0% [   3/75] b.gno")
	assert.Contains(t, lines[2], "2.0% [   2/100] c.gno")
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
	t.Parallel()
	tests := []struct {
		name            string
		pkgPath         string
		file            string
		node            *mockNode
		initialCoverage *CoverageData
		expectedHits    map[string]map[int]int
	}{
		{
			name:    "Record coverage for new file and line",
			pkgPath: "testpkg",
			file:    "testfile.gno",
			node: &mockNode{
				line:   10,
				column: 5,
			},
			initialCoverage: &CoverageData{
				Files: map[string]FileCoverage{
					"testpkg/testfile.gno": {
						HitLines:        make(map[int]int),
						ExecutableLines: map[int]bool{10: true}, // Add this line
					},
				},
				PkgPath:        "testpkg",
				CurrentPackage: "testpkg",
				CurrentFile:    "testfile.gno",
			},
			expectedHits: map[string]map[int]int{
				"testpkg/testfile.gno": {10: 1},
			},
		},
		{
			name:    "Increment hit count for existing line",
			pkgPath: "testpkg",
			file:    "testfile.gno",
			node: &mockNode{
				line:   10,
				column: 5,
			},
			initialCoverage: &CoverageData{
				Files: map[string]FileCoverage{
					"testpkg/testfile.gno": {
						HitLines:        map[int]int{10: 1},
						ExecutableLines: map[int]bool{10: true},
					},
				},
				PkgPath:        "testpkg",
				CurrentPackage: "testpkg",
				CurrentFile:    "testfile.gno",
			},
			expectedHits: map[string]map[int]int{
				"testpkg/testfile.gno": {10: 2},
			},
		},
		{
			name:    "Do not record coverage for non-executable line",
			pkgPath: "testpkg",
			file:    "testfile.gno",
			node: &mockNode{
				line:   20,
				column: 5,
			},
			initialCoverage: &CoverageData{
				Files: map[string]FileCoverage{
					"testpkg/testfile.gno": {
						HitLines:        map[int]int{},
						ExecutableLines: map[int]bool{10: true},
					},
				},
				PkgPath:        "testpkg",
				CurrentPackage: "testpkg",
				CurrentFile:    "testfile.gno",
			},
			expectedHits: map[string]map[int]int{
				"testpkg/testfile.gno": {},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &Machine{
				Coverage: tt.initialCoverage,
			}

			loc := m.recordCoverage(tt.node)

			// Check if the returned location is correct
			assert.Equal(t, tt.pkgPath, loc.PkgPath)
			assert.Equal(t, tt.file, loc.File)
			assert.Equal(t, tt.node.line, loc.Line)
			assert.Equal(t, tt.node.column, loc.Column)

			// Check if the coverage data has been updated correctly
			for file, expectedHits := range tt.expectedHits {
				actualHits := m.Coverage.Files[file].HitLines
				assert.Equal(t, expectedHits, actualHits)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

	c := NewCoverageData(rootDir)

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
			actualPath, err := c.findAbsoluteFilePath(tt.filePath)

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

	covData := NewCoverageData(tempDir)

	// 1st run: search from file system
	path1, err := covData.findAbsoluteFilePath("example.gno")
	if err != nil {
		t.Fatalf("failed to find absolute file path: %v", err)
	}
	assert.Equal(t, testFilePath, path1)

	// 2nd run: use cache
	path2, err := covData.findAbsoluteFilePath("example.gno")
	if err != nil {
		t.Fatalf("failed to find absolute file path: %v", err)
	}

	assert.Equal(t, testFilePath, path2)
	if len(covData.pathCache) != 1 {
		t.Fatalf("expected 1 path in cache, got %d", len(covData.pathCache))
	}
}

func TestDetectExecutableLines(t *testing.T) {
	t.Parallel()
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
			name: "Invalid gno code",
			content: `
This is not valid Go code
It should result in an error`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := detectExecutableLines(tt.content)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindFuncs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []*FuncExtent
	}{
		{
			name: "Single function",
			content: `
package main

func TestFunc() {
	println("Hello, World!")
}`,
			expected: []*FuncExtent{
				{
					name:      "TestFunc",
					startLine: 4,
					startCol:  1,
					endLine:   6,
					endCol:    2,
				},
			},
		},
		{
			name: "No functions",
			content: `
package main

var x = 10
`,
			expected: nil,
		},
		{
			name: "Function without body",
			content: `
package main

func TestFunc();
`,
			expected: nil,
		},
		{
			name: "Multiple functions",
			content: `
package main

func Func1() {
	println("Func1")
}

func Func2() {
	println("Func2")
}
`,
			expected: []*FuncExtent{
				{
					name:      "Func1",
					startLine: 4,
					startCol:  1,
					endLine:   6,
					endCol:    2,
				},
				{
					name:      "Func2",
					startLine: 8,
					startCol:  1,
					endLine:   10,
					endCol:    2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "example.go")
			assert.NoError(t, err)
			defer os.Remove(tmpfile.Name())
			_, err = tmpfile.Write([]byte(tt.content))
			assert.NoError(t, err)
			assert.NoError(t, tmpfile.Close())

			got, err := findFuncs(tmpfile.Name())
			if err != nil {
				t.Fatalf("findFuncs failed: %v", err)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("findFuncs() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCoverageData_ReportFuncCoverage(t *testing.T) {
	tests := []struct {
		name       string
		coverage   *CoverageData
		funcFilter string
		expected   string
	}{
		{
			name: "Single file, single function, no filter",
			coverage: &CoverageData{
				Functions: map[string][]FuncCoverage{
					"file1.gno": {
						{Name: "func1", Covered: 5, Total: 10},
					},
				},
			},
			funcFilter: "",
			expected: `Function Coverage:
File file1.gno:
func1                   5/10 50.0%
`,
		},
		{
			name: "Single file, multiple functions, no filter",
			coverage: &CoverageData{
				Functions: map[string][]FuncCoverage{
					"file1.gno": {
						{Name: "func1", Covered: 5, Total: 10},
						{Name: "func2", Covered: 8, Total: 8},
					},
				},
			},
			funcFilter: "",
			expected: `Function Coverage:
File file1.gno:
func1                   5/10 50.0%
func2                   8/8 100.0%
`,
		},
		{
			name: "Multiple files, multiple functions, with filter",
			coverage: &CoverageData{
				Functions: map[string][]FuncCoverage{
					"file1.gno": {
						{Name: "func1", Covered: 5, Total: 10},
						{Name: "func2", Covered: 8, Total: 8},
					},
					"file2.gno": {
						{Name: "func3", Covered: 0, Total: 5},
					},
				},
			},
			funcFilter: "func1",
			expected: `Function Coverage:
File file1.gno:
func1                   5/10 50.0%
`,
		},
		{
			name: "Multiple files, multiple functions, with regex filter",
			coverage: &CoverageData{
				Functions: map[string][]FuncCoverage{
					"file1.gno": {
						{Name: "testFunc1", Covered: 5, Total: 10},
						{Name: "testFunc2", Covered: 8, Total: 8},
					},
					"file2.gno": {
						{Name: "otherFunc", Covered: 0, Total: 5},
					},
				},
			},
			funcFilter: "^test",
			expected: `Function Coverage:
File file1.gno:
testFunc1               5/10 50.0%
testFunc2               8/8 100.0%
`,
		},
		{
			name: "No coverage",
			coverage: &CoverageData{
				Functions: map[string][]FuncCoverage{},
			},
			funcFilter: "",
			expected:   "Function Coverage:\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			io := commands.NewTestIO()
			io.SetOut(nopCloser{Buffer: &buf})

			tt.coverage.ReportFuncCoverage(io, tt.funcFilter)

			actual := removeANSIEscapeCodes(buf.String())
			expected := removeANSIEscapeCodes(tt.expected)

			assert.Equal(t, expected, actual)
		})
	}
}

func removeANSIEscapeCodes(s string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansi.ReplaceAllString(s, "")
}
