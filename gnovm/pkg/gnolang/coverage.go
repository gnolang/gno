package gnolang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// CoverageData stores code coverage information
type CoverageData struct {
	Files   map[string]FileCoverage
	PkgPath string
	RootDir string
}

// FileCoverage stores coverage information for a single file
type FileCoverage struct {
	TotalLines int
	HitLines   map[int]int
}

func NewCoverageData(rootDir string) *CoverageData {
	return &CoverageData{
		Files:   make(map[string]FileCoverage),
		PkgPath: "",
		RootDir: rootDir,
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

func (c *CoverageData) AddFile(filePath string, totalLines int) {
	if isTestFile(filePath) {
		return
	}

	fileCoverage, exists := c.Files[filePath]
	if !exists {
		fileCoverage = FileCoverage{
			HitLines: make(map[int]int),
		}
	}

	fileCoverage.TotalLines = totalLines
	c.Files[filePath] = fileCoverage
}

func (c *CoverageData) Report() {
	fmt.Println("Coverage Results:")
	for file, coverage := range c.Files {
		if !isTestFile(file) && strings.Contains(file, c.PkgPath) {
			hitLines := len(coverage.HitLines)
			percentage := float64(hitLines) / float64(coverage.TotalLines) * 100
			fmt.Printf("%s: %.2f%% (%d/%d lines)\n", file, percentage, hitLines, coverage.TotalLines)
		}
	}
}

func (c *CoverageData) ColoredCoverage(filePath string) error {
	realPath := filepath.Join(c.RootDir, "examples", filePath)
	if isTestFile(filePath) || !strings.Contains(realPath, c.PkgPath) || !strings.HasSuffix(realPath, ".gno") {
		return nil
	}
	file, err := os.Open(realPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1

	fileCoverage, exists := c.Files[filePath]
	if !exists {
		return fmt.Errorf("no coverage data for file %s", filePath)
	}

	fmt.Printf("Coverage Results for %s:\n", filePath)
	for scanner.Scan() {
		line := scanner.Text()
		if _, covered := fileCoverage.HitLines[lineNumber]; covered {
			fmt.Printf("%s%4d: %s%s\n", colorGreen, lineNumber, line, colorReset)
		} else {
			fmt.Printf("%s%4d: %s%s\n", colorYellow, lineNumber, line, colorReset)
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
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

func isTestFile(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "_test.gno") || strings.HasSuffix(pkgPath, "_testing.gno") || strings.HasSuffix(pkgPath, "_filetest.gno")
}

type JSONCoverage struct {
	Files map[string]JSONFileCoverage `json:"files"`
}

type JSONFileCoverage struct {
	TotalLines int            `json:"total_lines"`
	HitLines   map[string]int `json:"hit_lines"`
}

func (c *CoverageData) ToJSON() ([]byte, error) {
	jsonCov := JSONCoverage{
		Files: make(map[string]JSONFileCoverage),
	}

	for file, coverage := range c.Files {
		hitLines := make(map[string]int)
		for line, count := range coverage.HitLines {
			hitLines[strconv.Itoa(line)] = count
		}

		jsonCov.Files[file] = JSONFileCoverage{
			TotalLines: coverage.TotalLines,
			HitLines:   hitLines,
		}
	}

	return json.MarshalIndent(jsonCov, "", "  ")
}

func (c *CoverageData) SaveJSON(fileName string) error {
	data, err := c.ToJSON()
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, data, 0o644)
}
