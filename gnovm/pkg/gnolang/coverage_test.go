package gnolang

import (
	"reflect"
	"testing"
)

func TestNewCoverageData(t *testing.T) {
	cd := NewCoverageData()
	if cd == nil {
		t.Error("NewCoverageData() returned nil")
	}
	if cd == nil || cd.Files == nil {
		t.Error("NewCoverageData() did not initialize Files map")
	}
}

func TestCoverageDataAddHit(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		line     int
		expected map[string]*FileCoverage
	}{
		{
			name: "Add hit to new file",
			file: "test.go",
			line: 10,
			expected: map[string]*FileCoverage{
				"test.go": {
					Statements: map[int]int{10: 1},
				},
			},
		},
		{
			name: "Add hit to existing file",
			file: "test.go",
			line: 10,
			expected: map[string]*FileCoverage{
				"test.go": {
					Statements: map[int]int{10: 2},
				},
			},
		},
		{
			name: "Add hit to new line in existing file",
			file: "test.go",
			line: 20,
			expected: map[string]*FileCoverage{
				"test.go": {
					Statements: map[int]int{10: 2, 20: 1},
				},
			},
		},
	}

	cd := NewCoverageData()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd.AddHit(tt.file, tt.line)
			if !reflect.DeepEqual(cd.Files, tt.expected) {
				t.Errorf("AddHit() = %v, want %v", cd.Files, tt.expected)
			}
		})
	}
}

func TestCoverageData(t *testing.T) {
	tests := []struct {
		name string
		hits []struct {
			file string
			line int
		}
		wantReport string
	}{
		{
			name: "Empty coverage",
			hits: []struct {
				file string
				line int
			}{},
			wantReport: `Coverage Report:
=================

Total Coverage:
  Statements: 0
  Covered:    0
  Coverage:   0.00%
`,
		},
		{
			name: "Single file, single line",
			hits: []struct {
				file string
				line int
			}{
				{"file1.go", 10},
			},
			wantReport: `Coverage Report:
=================

file1.go:
  Statements: 1
  Covered:    1
  Coverage:   100.00%

Total Coverage:
  Statements: 1
  Covered:    1
  Coverage:   100.00%
`,
		},
		{
			name: "Multiple files, multiple lines",
			hits: []struct {
				file string
				line int
			}{
				{"file1.go", 10},
				{"file1.go", 20},
				{"file1.go", 10},
				{"file2.go", 5},
				{"file2.go", 15},
			},
			wantReport: `Coverage Report:
=================

file1.go:
  Statements: 2
  Covered:    2
  Coverage:   100.00%

file2.go:
  Statements: 2
  Covered:    2
  Coverage:   100.00%

Total Coverage:
  Statements: 4
  Covered:    4
  Coverage:   100.00%
`,
		},
		{
			name: "Partial coverage",
			hits: []struct {
				file string
				line int
			}{
				{"file1.go", 10},
				{"file1.go", 20},
				{"file2.go", 5},
			},
			wantReport: `Coverage Report:
=================

file1.go:
  Statements: 2
  Covered:    2
  Coverage:   100.00%

file2.go:
  Statements: 1
  Covered:    1
  Coverage:   100.00%

Total Coverage:
  Statements: 3
  Covered:    3
  Coverage:   100.00%
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := NewCoverageData()
			for _, hit := range tt.hits {
				cd.AddHit(hit.file, hit.line)
			}
			got := cd.Report()
			if got != tt.wantReport {
				t.Errorf("CoverageData.Report() =\n%v\nwant:\n%v", got, tt.wantReport)
			}
		})
	}
}
