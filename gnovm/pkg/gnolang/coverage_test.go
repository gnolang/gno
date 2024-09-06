package gnolang

import (
	"bytes"
	"os"
	"testing"
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
			initialData:  NewCoverageData(),
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
			initialData:   NewCoverageData(),
			expectedTotal: 100,
		},
		{
			name:          "Do not add test file *_test.gno",
			pkgPath:       "file1_test.gno",
			totalLines:    100,
			initialData:   NewCoverageData(),
			expectedTotal: 0,
		},
		{
			name:          "Do not add test file *_testing.gno",
			pkgPath:       "file1_testing.gno",
			totalLines:    100,
			initialData:   NewCoverageData(),
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

func TestPrintResults(t *testing.T) {
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

			tt.initialData.PrintResults()

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
