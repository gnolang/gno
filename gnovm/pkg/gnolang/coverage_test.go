package gnolang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func (m *mockNode) assertNode()                                      {}
func (m *mockNode) String() string                                   { return "" }
func (m *mockNode) Copy() Node                                       { return &mockNode{} }
func (m *mockNode) GetLabel() Name                                   { return "mockNode" }
func (m *mockNode) SetLabel(n Name)                                  {}
func (m *mockNode) GetLine() int                                     { return m.line }
func (m *mockNode) SetLine(l int)                                    {}
func (m *mockNode) GetColumn() int                                   { return m.column }
func (m *mockNode) SetColumn(c int)                                  {}
func (m *mockNode) DelAttribute(key GnoAttribute)                    {}
func (m *mockNode) GetAttribute(key GnoAttribute) interface{}        { return nil }
func (m *mockNode) HasAttribute(key GnoAttribute) bool               { return false }
func (m *mockNode) SetAttribute(key GnoAttribute, value interface{}) {}

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
				Enabled: true,
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
				Enabled: true,
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
				Enabled: true,
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

func TestViewFilesE2E(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	files := map[string]string{
		"file1.gno": "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n",
		"file2.gno": "package utils\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0o644)
		assert.NoError(t, err)
	}

	coverage := NewCoverageData(tempDir)
	for name, content := range files {
		execLines, err := detectExecutableLines(content)
		assert.NoError(t, err)
		coverage.SetExecutableLines(name, execLines)
		coverage.addFile(name, len(strings.Split(content, "\n")))
		coverage.updateHit(name, 4)
	}

	var buf bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(nopCloser{Buffer: &buf})

	err := coverage.ViewFiles("", true, io)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "file1.gno")
	assert.Contains(t, output, "file2.gno")
	assert.Contains(t, output, "println(\"Hello\")")
	assert.Contains(t, output, "return a + b")
	assert.Contains(t, output, string(colorGreen))
	assert.Contains(t, output, string(colorWhite))

	buf.Reset()
	err = coverage.ViewFiles("file1", true, io)
	assert.NoError(t, err)
	output = buf.String()
	assert.Contains(t, output, "file1.gno")
	assert.NotContains(t, output, "file2.gno")

	err = coverage.ViewFiles("nonexistent", true, io)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no files found matching pattern")

	buf.Reset()
	err = coverage.ViewFiles("", false, io)
	assert.NoError(t, err)
	output = buf.String()
	assert.NotContains(t, output, string(colorOrange))
}

func TestFormatLineInfoE2E(t *testing.T) {
	t.Parallel()
	coverage := NewCoverageData("")

	tests := []struct {
		name       string
		lineNumber int
		line       string
		hitCount   int
		covered    bool
		executable bool
		showHits   bool
		want       string
	}{
		{
			name:       "Covered line with hits",
			lineNumber: 1,
			line:       "println(\"Hello\")",
			hitCount:   2,
			covered:    true,
			executable: true,
			showHits:   true,
			want: fmt.Sprintf(
				"%s   1%s %s2   %s %sprintln(\"Hello\")%s",
				colorGreen, colorReset, colorOrange, colorReset, colorGreen, colorReset,
			),
		},
		{
			name:       "Executable but not covered line",
			lineNumber: 2,
			line:       "if x > 0 {",
			hitCount:   0,
			covered:    false,
			executable: true,
			showHits:   true,
			want: fmt.Sprintf(
				"%s   2%s      %sif x > 0 {%s",
				colorYellow, colorReset, colorYellow, colorReset,
			),
		},
		{
			name:       "Non-executable line",
			lineNumber: 3,
			line:       "package main",
			hitCount:   0,
			covered:    false,
			executable: false,
			showHits:   true,
			want: fmt.Sprintf(
				"%s   3%s      %spackage main%s",
				colorWhite, colorReset, colorWhite, colorReset,
			),
		},
		{
			name:       "Covered line without showing hits",
			lineNumber: 4,
			line:       "return x + y",
			hitCount:   1,
			covered:    true,
			executable: true,
			showHits:   false,
			want: fmt.Sprintf(
				"%s   4%s %sreturn x + y%s",
				colorGreen, colorReset, colorGreen, colorReset,
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := coverage.formatLineInfo(
				tt.lineNumber,
				tt.line,
				tt.hitCount,
				tt.covered,
				tt.executable,
				tt.showHits,
			)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindMatchingFilesE2E(t *testing.T) {
	t.Parallel()
	coverage := &CoverageData{
		Files: map[string]FileCoverage{
			"file1.gno":     {},
			"file2.gno":     {},
			"other_file.go": {},
		},
	}

	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "Match all .gno files",
			pattern: ".gno",
			want:    []string{"file1.gno", "file2.gno"},
		},
		{
			name:    "Match specific file",
			pattern: "file1",
			want:    []string{"file1.gno"},
		},
		{
			name:    "Match non-existent pattern",
			pattern: "nonexistent",
			want:    []string{},
		},
		{
			name:    "Match all files",
			pattern: "",
			want:    []string{"file1.gno", "file2.gno", "other_file.go"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := coverage.findMatchingFiles(tt.pattern)
			assert.ElementsMatch(t, tt.want, got)
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

func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, name)
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to create temp file %s: %v", filePath, err)
	}
	return filePath
}

func readFileContent(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(data)
}

func TestSaveHTML(t *testing.T) {
	tempDir := t.TempDir()

	source1 := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	source2 := `package utils

func Add(a, b int) int {
	return a + b
}`

	file1 := createTempFile(t, tempDir, "main.gno", source1)
	file2 := createTempFile(t, tempDir, "utils.gno", source2)

	coverage := NewCoverageData(tempDir)

	execLines1, err := detectExecutableLines(source1)
	if err != nil {
		t.Fatalf("Failed to detect executable lines for %s: %v", file1, err)
	}
	execLines2, err := detectExecutableLines(source2)
	if err != nil {
		t.Fatalf("Failed to detect executable lines for %s: %v", file2, err)
	}

	// Set executable lines
	relPath1, err := filepath.Rel(tempDir, file1)
	if err != nil {
		t.Fatalf("Failed to get relative path for %s: %v", file1, err)
	}
	relPath2, err := filepath.Rel(tempDir, file2)
	if err != nil {
		t.Fatalf("Failed to get relative path for %s: %v", file2, err)
	}
	coverage.SetExecutableLines(relPath1, execLines1)
	coverage.SetExecutableLines(relPath2, execLines2)

	// Add files with total executable lines
	totalExecLines1 := len(execLines1)
	totalExecLines2 := len(execLines2)
	coverage.addFile(relPath1, totalExecLines1)
	coverage.addFile(relPath2, totalExecLines2)

	// Simulate hits
	coverage.updateHit(relPath1, 6) // fmt.Println line
	coverage.updateHit(relPath2, 4) // return a + b

	// Define output HTML file
	outputHTML := filepath.Join(tempDir, "coverage.html")

	// Run SaveHTML
	err = coverage.SaveHTML(outputHTML)
	if err != nil {
		t.Fatalf("SaveHTML failed: %v", err)
	}

	// Read and verify the HTML content
	htmlContent := readFileContent(t, outputHTML)

	// Basic checks
	if !strings.Contains(htmlContent, "main.gno") {
		t.Errorf("HTML does not contain main.gno")
	}
	if !strings.Contains(htmlContent, "utils.gno") {
		t.Errorf("HTML does not contain utils.gno")
	}

	// Check for hit counts
	if !strings.Contains(htmlContent, ">1</span>") {
		t.Errorf("Expected hit count '1' for main.gno, but not found")
	}
	if !strings.Contains(htmlContent, ">1</span>") {
		t.Errorf("Expected hit count '1' for utils.gno, but not found")
	}
}

func TestSaveHTML_EmptyCoverage(t *testing.T) {
	tempDir := t.TempDir()

	coverage := NewCoverageData(tempDir)

	outputHTML := filepath.Join(tempDir, "coverage_empty.html")

	err := coverage.SaveHTML(outputHTML)
	if err != nil {
		t.Fatalf("SaveHTML failed: %v", err)
	}

	htmlContent := readFileContent(t, outputHTML)
	if !strings.Contains(htmlContent, "<h1>Coverage Report</h1>") {
		t.Errorf("HTML does not contain the main heading")
	}
	if strings.Contains(htmlContent, "class=\"filename\"") {
		t.Errorf("HTML should not contain any filenames, but found some")
	}
}

func TestSaveHTML_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	sources := map[string]string{
		"file1.gno": `package a

func A() {
	// Line 3
}`,
		"file2.gno": `package b

func B() {
	// Line 3
	// Line 4
}`,
		"file3.gno": `package c

func C() {
	// Line 3
	// Line 4
	// Line 5
}`,
	}

	for name, content := range sources {
		createTempFile(t, tempDir, name, content)
	}

	coverage := NewCoverageData(tempDir)

	for name, content := range sources {
		relPath := name
		execLines, err := detectExecutableLines(content)
		if err != nil {
			t.Fatalf("Failed to detect executable lines for %s: %v", name, err)
		}
		coverage.SetExecutableLines(relPath, execLines)
		totalExecLines := len(execLines)
		coverage.addFile(relPath, totalExecLines)
	}

	coverage.updateHit("file1.gno", 3) // Line 3
	coverage.updateHit("file2.gno", 3) // Line 3
	coverage.updateHit("file2.gno", 4) // Line 4
	coverage.updateHit("file3.gno", 3) // Line 3

	outputHTML := filepath.Join(tempDir, "coverage_multiple.html")

	err := coverage.SaveHTML(outputHTML)
	if err != nil {
		t.Fatalf("SaveHTML failed: %v", err)
	}

	htmlContent := readFileContent(t, outputHTML)

	for name := range sources {
		if !strings.Contains(htmlContent, name) {
			t.Errorf("HTML does not contain %s", name)
		}
	}
}

func TestSaveHTML_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	coverage := NewCoverageData(tempDir)
	coverage.SetExecutableLines("nonexistent.gno", map[int]bool{1: true, 2: true})
	coverage.addFile("nonexistent.gno", 2)
	coverage.updateHit("nonexistent.gno", 1)

	outputHTML := filepath.Join(tempDir, "coverage_notfound.html")

	err := coverage.SaveHTML(outputHTML)
	if err == nil {
		t.Fatalf("Expected SaveHTML to fail due to missing file, but it succeeded")
	}

	if !strings.Contains(err.Error(), "file nonexistent.gno not found") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestSaveHTML_FileCreationError(t *testing.T) {
	tempDir := t.TempDir()

	createTempFile(t, tempDir, "file.gno", `package main

func main() {}`)

	coverage := NewCoverageData(tempDir)
	coverage.Files["file.gno"] = FileCoverage{
		TotalLines:      2, // Assuming 2 executable lines
		HitLines:        map[int]int{1: 1},
		ExecutableLines: map[int]bool{1: true, 2: true},
	}

	outputHTML := tempDir

	err := coverage.SaveHTML(outputHTML)
	if err == nil {
		t.Fatalf("Expected SaveHTML to fail due to invalid file path, but it succeeded")
	}

	if !strings.Contains(err.Error(), "is a directory") {
		t.Errorf("Unexpected error message: %v", err)
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
