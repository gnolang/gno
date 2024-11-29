package coverage

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsoleReporter_WriteReport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupCoverage func() *Coverage
		wantContains  []string
		wantExcludes  []string
	}{
		{
			name: "basic coverage report",
			setupCoverage: func() *Coverage {
				cov := New("")
				cov.Enable()

				filePath := "main/file.gno"
				execLines := map[int]bool{
					4: true,
					5: true,
				}
				cov.SetExecutableLines(filePath, execLines)
				cov.AddFile(filePath, 10)
				cov.RecordHit(FileLocation{
					PkgPath: "main",
					File:    "file.gno",
					Line:    4,
				})

				return cov
			},
			wantContains: []string{
				"50.0%",
				"1/2",
				"file.gno",
				string(Yellow),
			},
		},
		{
			name: "high coverage report",
			setupCoverage: func() *Coverage {
				cov := New("")
				cov.Enable()

				filePath := "pkg/high.gno"
				execLines := map[int]bool{
					1: true,
					2: true,
				}
				cov.SetExecutableLines(filePath, execLines)
				cov.AddFile(filePath, 10)

				cov.RecordHit(FileLocation{PkgPath: "pkg", File: "high.gno", Line: 1})
				cov.RecordHit(FileLocation{PkgPath: "pkg", File: "high.gno", Line: 2})

				return cov
			},
			wantContains: []string{
				"100.0%",
				"2/2",
				"high.gno",
				string(Green),
			},
		},
		{
			name: "low coverage report",
			setupCoverage: func() *Coverage {
				cov := New("")
				cov.Enable()

				filePath := "pkg/low.gno"
				execLines := map[int]bool{
					1: true,
					2: true,
					3: true,
					4: true,
					5: true,
				}
				cov.SetExecutableLines(filePath, execLines)
				cov.AddFile(filePath, 10)

				cov.RecordHit(FileLocation{PkgPath: "pkg", File: "low.gno", Line: 1})

				return cov
			},
			wantContains: []string{
				"20.0%",
				"1/5",
				"low.gno",
				string(Red),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cov := tt.setupCoverage()
			reporter := NewConsoleReporter(cov, NewDefaultPathFinder(""))

			err := reporter.Write(&buf)
			require.NoError(t, err)

			output := buf.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
			for _, exclude := range tt.wantExcludes {
				assert.NotContains(t, output, exclude)
			}
		})
	}
}

func TestConsoleReporter_WriteFileDetail(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFileName := "test.gno"
	testPath := filepath.Join(tempDir, testFileName)
	testContent := `package test

func Add(a, b int) int {
    return a + b
}
`
	require.NoError(t, os.WriteFile(testPath, []byte(testContent), 0o644))

	tests := []struct {
		name          string
		pattern       string
		showHits      bool
		setupCoverage func() *Coverage
		wantContains  []string
		wantErr       bool
	}{
		{
			name:     "show file with hits",
			pattern:  testFileName,
			showHits: true,
			setupCoverage: func() *Coverage {
				cov := New(tempDir)
				cov.Enable()

				execLines, _ := DetectExecutableLines(testContent)
				cov.SetExecutableLines(testFileName, execLines)
				cov.AddFile(testFileName, len(strings.Split(testContent, "\n")))
				cov.RecordHit(FileLocation{File: testFileName, Line: 4})

				return cov
			},
			wantContains: []string{
				testFileName,
				"func Add",
				"return a + b",
				string(Green),  // covered line
				string(White),  // non-executable line
				string(Orange), // hit count
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cov := tt.setupCoverage()
			reporter := NewConsoleReporter(cov, NewDefaultPathFinder(tempDir))

			err := reporter.WriteFileDetail(&buf, tt.pattern, tt.showHits)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

func TestJSONReporter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupCoverage func() *Coverage
		checkOutput   func(*testing.T, []byte)
	}{
		{
			name: "basic json report",
			setupCoverage: func() *Coverage {
				cov := New("")
				cov.Enable()

				filePath := "pkg/file.gno"
				cov.AddFile(filePath, 10)
				cov.SetExecutableLines(filePath, map[int]bool{1: true})
				cov.RecordHit(FileLocation{PkgPath: "pkg", File: "file.gno", Line: 1})

				return cov
			},
			checkOutput: func(t *testing.T, output []byte) {
				var report jsonCoverage
				require.NoError(t, json.Unmarshal(output, &report))

				assert.Contains(t, report.Files, "pkg/file.gno")
				fileCov := report.Files["pkg/file.gno"]
				assert.Equal(t, 10, fileCov.TotalLines)
				assert.Equal(t, 1, fileCov.HitLines["1"])
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cov := tt.setupCoverage()
			reporter := NewJSONReporter(cov, "")

			err := reporter.Write(&buf)
			require.NoError(t, err)

			tt.checkOutput(t, buf.Bytes())

			println(buf.String())
		})
	}
}
