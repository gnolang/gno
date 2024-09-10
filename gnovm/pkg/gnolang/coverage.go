package gnolang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
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

	// Sort files by name for consistent output
	var files []string
	for file := range c.Files {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		coverage := c.Files[file]
		if !isTestFile(file) && strings.Contains(file, c.PkgPath) {
			hitLines := len(coverage.HitLines)
			percentage := float64(hitLines) / float64(coverage.TotalLines) * 100
			fmt.Printf("%s: %.2f%% (%d/%d lines)\n", file, percentage, hitLines, coverage.TotalLines)
		}
	}
}

func (c *CoverageData) ColoredCoverage(filePath string) error {
	realPath, err := c.determineRealPath(filePath)
	if err != nil {
		// skipping invalid file paths
		return nil
	}

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

// Attempts to determine the full real path based on the filePath alone.
// It dynamically checks if the file exists in either examples or gnovm/stdlibs directories.
func (c *CoverageData) determineRealPath(filePath string) (string, error) {
	if !strings.HasSuffix(filePath, ".gno") {
		return "", fmt.Errorf("invalid file type: %s (not a .gno file)", filePath)
	}
	if isTestFile(filePath) {
		return "", fmt.Errorf("cannot determine real path for test file: %s", filePath)
	}

	// Define possible base directories
	baseDirs := []string{
		filepath.Join(c.RootDir, "examples"), // p, r packages
		filepath.Join(c.RootDir, "gnovm", "stdlibs"),
	}

	// Try finding the file in each base directory
	for _, baseDir := range baseDirs {
		realPath := filepath.Join(baseDir, filePath)

		// Check if the file exists
		if _, err := os.Stat(realPath); err == nil {
			return realPath, nil
		}
	}

	return "", fmt.Errorf("file %s not found in known paths", filePath)
}

func isTestFile(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "_test.gno") ||
		strings.HasSuffix(pkgPath, "_testing.gno") ||
		strings.HasSuffix(pkgPath, "_filetest.gno")
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

func (m *Machine) AddFileToCodeCoverage(file string, totalLines int) {
	if isTestFile(file) {
		return
	}
	m.Coverage.AddFile(file, totalLines)
}

// recordCoverage records the execution of a specific node in the AST.
// This function tracking which parts of the code have been executed during the runtime.
//
// Note: This function assumes that CurrentPackage and CurrentFile are correctly set in the Machine
// before it's called. These fields provide the context necessary to accurately record the coverage information.
func (m *Machine) recordCoverage(node Node) Location {
	if node == nil {
		return Location{}
	}

	pkgPath := m.CurrentPackage
	file := m.CurrentFile
	line := node.GetLine()

	path := filepath.Join(pkgPath, file)
	m.Coverage.AddHit(path, line)

	return Location{
		PkgPath: pkgPath,
		File:    file,
		Line:    line,
		Column:  node.GetColumn(),
	}
}

// region Executable Lines Detection

func countCodeLines(content string) int {
	lines, err := detectExecutableLines(content)
	if err != nil {
		return 0
	}

	return len(lines)
}

// TODO: use gno Node type
func isExecutableLine(node ast.Node) bool {
	switch node.(type) {
	case *ast.AssignStmt, *ast.ExprStmt, *ast.ReturnStmt, *ast.BranchStmt,
		*ast.IncDecStmt, *ast.GoStmt, *ast.DeferStmt, *ast.SendStmt:
		return true
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt:
		return true
	case *ast.FuncDecl:
		return false
	case *ast.BlockStmt:
		return false
	case *ast.DeclStmt:
		return false
	case *ast.ImportSpec, *ast.TypeSpec, *ast.ValueSpec:
		return false
	case *ast.GenDecl:
		return false
	default:
		return false
	}
}

func detectExecutableLines(content string) (map[int]bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	executableLines := make(map[int]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		if isExecutableLine(n) {
			line := fset.Position(n.Pos()).Line
			executableLines[line] = true
		}

		return true
	})

	return executableLines, nil
}
