package gnolang

import (
	"fmt"
	"strings"
)

// CoverageData stores code coverage information
type CoverageData struct {
	Files map[string]FileCoverage
}

// FileCoverage stores coverage information for a single file
type FileCoverage struct {
	TotalLines int
	HitLines   map[int]int
}

func NewCoverageData() *CoverageData {
	return &CoverageData{
		Files: make(map[string]FileCoverage),
	}
}

func (c *CoverageData) AddHit(pkgPath string, line int) {
	fileCoverage, exists := c.Files[pkgPath]
	if !exists {
		fileCoverage = FileCoverage{
			TotalLines: 0,
			HitLines:   make(map[int]int),
		}
		c.Files[pkgPath] = fileCoverage
	}

	fileCoverage.HitLines[line]++

	// Only update the file coverage, without incrementing TotalLines
	c.Files[pkgPath] = fileCoverage
}

func isTestFile(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "_test.gno") || strings.HasSuffix(pkgPath, "_testing.gno")
}

func (c *CoverageData) AddFile(pkgPath string, totalLines int) {
	if isTestFile(pkgPath) {
        return
    }

	fileCoverage, exists := c.Files[pkgPath]
	if !exists {
		fileCoverage = FileCoverage{
			HitLines: make(map[int]int),
		}
	}

	fileCoverage.TotalLines = totalLines
	c.Files[pkgPath] = fileCoverage
}

func (c *CoverageData) PrintResults() {
    fmt.Println("Coverage Results:")
    for file, coverage := range c.Files {
		if !isTestFile(file) {
			hitLines := len(coverage.HitLines)
			percentage := float64(hitLines) / float64(coverage.TotalLines) * 100
			fmt.Printf("%s: %.2f%% (%d/%d lines)\n", file, percentage, hitLines, coverage.TotalLines)
		}
    }
}

func countCodeLines(content string) int {
	lines := strings.Split(content, "\n")
	codeLines := 0
	inBlockComment := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if inBlockComment {
			if strings.Contains(trimmedLine, "*/") {
				inBlockComment = false
			}
			continue
		}

		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		if strings.HasPrefix(trimmedLine, "/*") {
            inBlockComment = true
            if strings.Contains(trimmedLine, "*/") {
                inBlockComment = false
            }
            continue
        }

        codeLines++
	}

	return codeLines
}
