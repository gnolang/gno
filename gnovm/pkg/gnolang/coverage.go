package gnolang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

const (
	colorReset  = "\033[0m"
	colorOrange = "\033[38;5;208m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorWhite  = "\033[37m"
	boldText    = "\033[1m"
)

// CoverageData stores code coverage information
type CoverageData struct {
	Files          map[string]FileCoverage
	PkgPath        string
	RootDir        string
	CurrentPackage string
	CurrentFile    string
}

// FileCoverage stores coverage information for a single file
type FileCoverage struct {
	TotalLines      int
	HitLines        map[int]int
	ExecutableLines map[int]bool
}

func NewCoverageData(rootDir string) *CoverageData {
	return &CoverageData{
		Files:          make(map[string]FileCoverage),
		PkgPath:        "",
		RootDir:        rootDir,
		CurrentPackage: "",
		CurrentFile:    "",
	}
}

func (c *CoverageData) SetExecutableLines(filePath string, executableLines map[int]bool) {
	cov, exists := c.Files[filePath]
	if !exists {
		cov = FileCoverage{
			TotalLines:      0,
			HitLines:        make(map[int]int),
			ExecutableLines: make(map[int]bool),
		}
	}

	cov.ExecutableLines = executableLines
	c.Files[filePath] = cov
}

func (c *CoverageData) AddHit(pkgPath string, line int) {
	if !strings.HasSuffix(pkgPath, ".gno") {
		return
	}
	if isTestFile(pkgPath) {
		return
	}

	fileCoverage, exists := c.Files[pkgPath]
	if !exists {
		fileCoverage = FileCoverage{
			TotalLines: 0,
			HitLines:   make(map[int]int),
		}
		c.Files[pkgPath] = fileCoverage
	}

	if fileCoverage.ExecutableLines[line] {
		fileCoverage.HitLines[line]++
	}

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

// region Reporting

func (c *CoverageData) ViewFiles(pattern string, showHits bool, io commands.IO) error {
	matchingFiles := c.findMatchingFiles(pattern)
	if len(matchingFiles) == 0 {
		return fmt.Errorf("no files found matching pattern %s", pattern)
	}

	for _, path := range matchingFiles {
		err := c.viewSingleFileCoverage(path, showHits, io)
		if err != nil {
			return err
		}
		io.Println() // Add a newline between files
	}

	return nil
}

func (c *CoverageData) viewSingleFileCoverage(filePath string, showHits bool, io commands.IO) error {
	realPath, err := c.determineRealPath(filePath)
	if err != nil {
		// skipping invalid file paths
		return nil
	}

	file, err := os.Open(realPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1
	coverage, exists := c.Files[filePath]
	if !exists {
		return fmt.Errorf("no coverage data for file %s", filePath)
	}

	io.Printfln("%s%s%s:", boldText, filePath, colorReset)
	for scanner.Scan() {
		line := scanner.Text()
		hitCount, covered := coverage.HitLines[lineNumber]

		var hitInfo string
		if showHits {
			if covered {
				hitInfo = fmt.Sprintf("%s%d%s ", colorOrange, hitCount, colorReset)
			} else {
				hitInfo = strings.Repeat(" ", 2)
			}
		}

		lineNumStr := fmt.Sprintf("%4d", lineNumber)

		if showHits {
			if covered {
				io.Printfln("%s%s%s %-4s %s%s%s", colorGreen, lineNumStr, colorReset, hitInfo, colorGreen, line, colorReset)
			} else if coverage.ExecutableLines[lineNumber] {
				io.Printfln("%s%s%s %-4s %s%s%s", colorYellow, lineNumStr, colorReset, hitInfo, colorYellow, line, colorReset)
			} else {
				io.Printfln("%s%s%s %-4s %s%s", colorWhite, lineNumStr, colorReset, hitInfo, line, colorReset)
			}
		} else {
			if covered {
				io.Printfln("%s%s %s%s", colorGreen, lineNumStr, line, colorReset)
			} else if coverage.ExecutableLines[lineNumber] {
				io.Printfln("%s%s %s%s", colorYellow, lineNumStr, line, colorReset)
			} else {
				io.Printfln("%s%s %s%s", colorWhite, lineNumStr, line, colorReset)
			}
		}
		lineNumber++
	}

	return scanner.Err()
}

func (c *CoverageData) findMatchingFiles(pattern string) []string {
	var files []string
	for file := range c.Files {
		if strings.Contains(file, pattern) {
			files = append(files, file)
		}
	}
	return files
}

func (c *CoverageData) ListFiles(io commands.IO) {
	for file, cov := range c.Files {
		hitLines := len(cov.HitLines)
		totalLines := cov.TotalLines
		pct := float64(hitLines) / float64(totalLines) * 100
		color := getCoverageColor(pct)
		io.Printfln("%s%3.0f%% [%d/%d] %s%s", color, pct, hitLines, totalLines, file, colorReset)
	}
}

func getCoverageColor(percentage float64) string {
	switch {
	case percentage >= 80:
		return colorGreen
	case percentage >= 50:
		return colorYellow
	default:
		return colorRed
	}
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

// Attempts to determine the full real path based on the filePath alone.
// It dynamically checks if the file exists in either examples or gnovm/stdlibs directories.
func (c *CoverageData) determineRealPath(filePath string) (string, error) {
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

func (c *CoverageData) SaveHTML(outputFileName string) error {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Coverage Report</title>
    <style>
        body { font-family: 'Courier New', Courier, monospace; line-height: 1.5; }
        .file { margin-bottom: 20px; }
        .filename { font-weight: bold; margin-bottom: 10px; }
        pre { margin: 0; }
        .line { display: flex; }
        .line-number { color: #999; padding-right: 1em; text-align: right; width: 4em; }
        .hit-count { color: #666; padding-right: 1em; width: 3em; text-align: right; }
        .covered { background-color: #90EE90; }
        .uncovered { background-color: #FFB6C1; }
    </style>
</head>
<body>
    <h1>Coverage Report</h1>
    {{range $file, $coverage := .Files}}
    <div class="file">
        <div class="filename">{{$file}}</div>
        <pre>{{range $line, $content := $coverage.Lines}}
<span class="line{{if $content.Covered}} covered{{else if $content.Executable}} uncovered{{end}}"><span class="line-number">{{$line}}</span><span class="hit-count">{{if $content.Covered}}{{$content.Hits}}{{else}}-{{end}}</span><span class="code">{{$content.Code}}</span></span>{{end}}
        </pre>
    </div>
    {{end}}
</body>
</html>`

	t, err := template.New("coverage").Parse(tmpl)
	if err != nil {
		return err
	}

	data := struct {
		Files map[string]struct {
			Lines map[int]struct {
				Code       string
				Covered    bool
				Executable bool
				Hits       int
			}
		}
	}{
		Files: make(map[string]struct {
			Lines map[int]struct {
				Code       string
				Covered    bool
				Executable bool
				Hits       int
			}
		}),
	}

	for path, coverage := range c.Files {
		realPath, err := c.determineRealPath(path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(realPath)
		if err != nil {
			return err
		}

		lines := strings.Split(string(content), "\n")
		fileData := struct {
			Lines map[int]struct {
				Code       string
				Covered    bool
				Executable bool
				Hits       int
			}
		}{
			Lines: make(map[int]struct {
				Code       string
				Covered    bool
				Executable bool
				Hits       int
			}),
		}

		for i, line := range lines {
			lineNum := i + 1
			hits, covered := coverage.HitLines[lineNum]
			executable := coverage.ExecutableLines[lineNum]

			fileData.Lines[lineNum] = struct {
				Code       string
				Covered    bool
				Executable bool
				Hits       int
			}{
				Code:       line,
				Covered:    covered,
				Executable: executable,
				Hits:       hits,
			}
		}

		data.Files[path] = fileData
	}

	file, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
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

	pkgPath := m.Coverage.CurrentPackage
	file := m.Coverage.CurrentFile
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

// countCodeLines counts the number of executable lines in the given source code content.
func countCodeLines(content string) int {
	lines, err := detectExecutableLines(content)
	if err != nil {
		return 0
	}

	return len(lines)
}

// isExecutableLine determines whether a given AST node represents an
// executable line of code for the purpose of code coverage measurement.
//
// It returns true for statement nodes that typically contain executable code,
// such as assignments, expressions, return statements, and control flow statements.
//
// It returns false for nodes that represent non-executable lines, such as
// declarations, blocks, and function definitions.
func isExecutableLine(node ast.Node) bool {
	switch node.(type) {
	case *ast.AssignStmt, *ast.ExprStmt, *ast.ReturnStmt, *ast.BranchStmt,
		*ast.IncDecStmt, *ast.GoStmt, *ast.DeferStmt, *ast.SendStmt:
		return true
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt:
		return true
	case *ast.CaseClause:
		// Even if a `case` condition (e.g., `case 1:`) in a `switch` statement is executed,
		// the condition itself is not included in the coverage; coverage only recorded for the
		// code block inside the corresponding `case` clause.
		return false
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
