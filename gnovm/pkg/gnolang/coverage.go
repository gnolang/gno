package gnolang

import (
	"fmt"
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
			HitLines: make(map[int]int),
		}
	}

	fileCoverage.TotalLines++
	fileCoverage.HitLines[line]++
	c.Files[pkgPath] = fileCoverage
}

func (c *CoverageData) AddFile(pkgPath string, totalLines int) {
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
        hitLines := len(coverage.HitLines)
        percentage := float64(hitLines) / float64(coverage.TotalLines) * 100
        fmt.Printf("%s: %.2f%% (%d/%d lines)\n", file, percentage, hitLines, coverage.TotalLines)
    }
}
