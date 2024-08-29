package gnolang

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CoverageData struct {
	Files map[string]*FileCoverage
	SourceCode map[string][]string
	Debug bool
}

type FileCoverage struct {
	Statements map[int]int // line number -> execution count
	TotalLines int // total number of lines in the file
}

func NewCoverageData() *CoverageData {
	return &CoverageData{
		Files: make(map[string]*FileCoverage),
		SourceCode: make(map[string][]string),
		Debug: true,
	}
}

func (c *CoverageData) AddFile(file string, totalLines int) {
	if _, exists := c.Files[file]; !exists {
		c.Files[file] = &FileCoverage{
			TotalLines: totalLines,
			Statements: make(map[int]int),
		}
	}
}

func (c *CoverageData) LoadSourceCode(rootDir string) error {
	for file := range c.Files {
		fullPath := filepath.Join(rootDir, file)
		f, err := os.Open(fullPath)
		if err != nil {
			return err
		}
		defer f.Close()

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		c.SourceCode[file] = lines
		c.Files[file].TotalLines = len(lines)
	}

	return nil
}

func (c *CoverageData) AddHit(file string, line int) {
	if c.Files[file] == nil {
		c.AddFile(file, line)
	}

	c.Files[file].Statements[line]++
}

func (c *CoverageData) SetDebug(debug bool) {
	c.Debug = debug
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
        totalLines := fileCoverage.TotalLines
        stmts := len(fileCoverage.Statements)
		covered := 0

		var coveredLines []int
		for line, count := range fileCoverage.Statements {
			if count > 0 {
				covered++
				coveredLines = append(coveredLines, line)
			}
		}

		sort.Ints(coveredLines)

        percentage := float64(stmts) / float64(totalLines) * 100
        report.WriteString(fmt.Sprintf("%s:\n", fileName))
        report.WriteString(fmt.Sprintf("  Total Lines: %d\n", totalLines))
        report.WriteString(fmt.Sprintf("  Covered:     %d\n", stmts))
        report.WriteString(fmt.Sprintf("  Coverage:    %.2f%%\n\n", percentage))

        totalStatements += totalLines
        totalCovered += stmts

		if c.Debug {
			report.WriteString("  Covered lines: ")
            for i, line := range coveredLines {
                if i > 0 {
                    report.WriteString(", ")
                }
                report.WriteString(fmt.Sprintf("%d", line))
            }
            report.WriteString("\n")
		}

		report.WriteString("\n")
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
