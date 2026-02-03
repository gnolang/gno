package coverage

import (
	"fmt"
	"sort"
	"strings"
)

// FileCoverage represents coverage data for a single file.
type FileCoverage struct {
	Package         string
	FileName        string
	ExecutableLines []int
	ExecutedLines   map[int]int64
	TotalLines      int
	CoveredLines    int
	Coverage        float64
}

// Report represents a complete coverage report.
type Report struct {
	Files        []*FileCoverage
	TotalLines   int
	CoveredLines int
	Coverage     float64
}

// GenerateReport creates a coverage report from the tracker data.
func (t *Tracker) GenerateReport() *Report {
	t.mu.RLock()
	defer t.mu.RUnlock()

	report := &Report{
		Files: make([]*FileCoverage, 0),
	}

	// Process each package and file
	for pkgPath, pkgFiles := range t.executableLines {
		for fileName, execLines := range pkgFiles {
			fc := &FileCoverage{
				Package:         pkgPath,
				FileName:        fileName,
				ExecutableLines: make([]int, 0, len(execLines)),
				ExecutedLines:   make(map[int]int64),
			}

			// Collect executable lines
			for line := range execLines {
				fc.ExecutableLines = append(fc.ExecutableLines, line)
			}
			sort.Ints(fc.ExecutableLines)
			fc.TotalLines = len(fc.ExecutableLines)

			// Collect executed lines
			if covData, ok := t.coverage[pkgPath]; ok {
				if fileCov, ok := covData[fileName]; ok {
					for line, count := range fileCov {
						if execLines[line] { // Only count if line is executable
							fc.ExecutedLines[line] = count
							fc.CoveredLines++
						}
					}
				}
			}

			// Calculate coverage percentage
			if fc.TotalLines > 0 {
				fc.Coverage = float64(fc.CoveredLines) / float64(fc.TotalLines) * 100
			}

			report.Files = append(report.Files, fc)
			report.TotalLines += fc.TotalLines
			report.CoveredLines += fc.CoveredLines
		}
	}

	// Sort files by package and name
	sort.Slice(report.Files, func(i, j int) bool {
		if report.Files[i].Package != report.Files[j].Package {
			return report.Files[i].Package < report.Files[j].Package
		}
		return report.Files[i].FileName < report.Files[j].FileName
	})

	// Calculate overall coverage
	if report.TotalLines > 0 {
		report.Coverage = float64(report.CoveredLines) / float64(report.TotalLines) * 100
	}

	return report
}

// String returns a text representation of the coverage report.
func (r *Report) String() string {
	var sb strings.Builder

	sb.WriteString("Coverage Report\n")
	sb.WriteString("===============\n\n")

	for _, file := range r.Files {
		sb.WriteString(fmt.Sprintf("%s/%s: %.1f%% (%d/%d lines)\n",
			file.Package, file.FileName, file.Coverage,
			file.CoveredLines, file.TotalLines))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Total Coverage: %.1f%% (%d/%d lines)\n",
		r.Coverage, r.CoveredLines, r.TotalLines))

	return sb.String()
}

// GetUncoveredLines returns lines that were not executed for a specific file.
func (fc *FileCoverage) GetUncoveredLines() []int {
	uncovered := make([]int, 0)

	for _, line := range fc.ExecutableLines {
		if _, executed := fc.ExecutedLines[line]; !executed {
			uncovered = append(uncovered, line)
		}
	}

	return uncovered
}
