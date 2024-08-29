package gnolang

import (
	"fmt"
	"sort"
	"strings"
)

type CoverageData struct {
	Files map[string]*FileCoverage
}

type FileCoverage struct {
	Statements map[int]int // line number -> count
}

func NewCoverageData() *CoverageData {
	return &CoverageData{
		Files: make(map[string]*FileCoverage),
	}
}

func (c *CoverageData) AddHit(file string, line int) {
	if c.Files[file] == nil {
		c.Files[file] = &FileCoverage{
			Statements: make(map[int]int),
		}
	}

	c.Files[file].Statements[line]++
}

func (c *CoverageData) Report() string {
	var report strings.Builder
	report.WriteString("Coverage Report:\n")
	report.WriteString("=================\n\n")

	var fileNames []string
	for fileName := range c.Files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	totalStatements := 0
	totalCovered := 0

	for _, fileName := range fileNames {
		fileCoverage := c.Files[fileName]
		statements := len(fileCoverage.Statements)
		covered := 0
		for _, count := range fileCoverage.Statements {
			if count > 0 {
				covered++
			}
		}

		totalStatements += statements
		totalCovered += covered

		percentage := calculateCoverage(covered, statements)
		report.WriteString(fmt.Sprintf("%s:\n", fileName))
		report.WriteString(fmt.Sprintf("  Statements: %d\n", statements))
		report.WriteString(fmt.Sprintf("  Covered:    %d\n", covered))
		report.WriteString(fmt.Sprintf("  Coverage:   %.2f%%\n\n", percentage))
	}

	totalPercentage := calculateCoverage(totalCovered, totalStatements)
	report.WriteString("Total Coverage:\n")
	report.WriteString(fmt.Sprintf("  Statements: %d\n", totalStatements))
	report.WriteString(fmt.Sprintf("  Covered:    %d\n", totalCovered))
	report.WriteString(fmt.Sprintf("  Coverage:   %.2f%%\n", totalPercentage))

	return report.String()
}

func calculateCoverage(a, b int) float64 {
	if a == 0 || b == 0 {
		return 0.0
	}
	return float64(a) / float64(b) * 100
}
