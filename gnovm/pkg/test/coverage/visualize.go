package coverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ANSI color codes
const (
	ColorReset = "\033[0m"
	ColorGreen = "\033[32m"
	ColorRed   = "\033[31m"
	ColorGray  = "\033[90m"
	ColorWhite = "\033[37m"
	ColorBold  = "\033[1m"
)

// ShowFileCoverage displays coverage visualization for a specific file
func (t *Tracker) ShowFileCoverage(rootDir, pattern string, w io.Writer) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Find files matching the pattern
	var matchedFiles []*FileCoverage
	report := t.GenerateReport()

	for _, fc := range report.Files {
		fileName := fc.FileName
		if matched, _ := filepath.Match(pattern, fileName); matched {
			matchedFiles = append(matchedFiles, fc)
		} else if matched, _ := filepath.Match(pattern, filepath.Base(fileName)); matched {
			matchedFiles = append(matchedFiles, fc)
		}
	}

	if len(matchedFiles) == 0 {
		return fmt.Errorf("no files matching pattern: %s", pattern)
	}

	// If multiple files matched, use interactive viewer
	if len(matchedFiles) > 1 {
		// Convert to file paths for interactive viewer
		var filePaths []string
		for _, fc := range matchedFiles {
			filePaths = append(filePaths, fmt.Sprintf("%s/%s", fc.Package, fc.FileName))
		}

		viewer := NewInteractiveViewer(t, rootDir, filePaths, os.Stdin, w)
		return viewer.Start()
	}

	// Single file - show directly
	return t.showSingleFileCoverage(rootDir, matchedFiles[0], w)
}

// showSingleFileCoverage displays coverage for a single file
func (t *Tracker) showSingleFileCoverage(rootDir string, fc *FileCoverage, w io.Writer) error {
	// Try to find the file
	filePath := findSourceFile(rootDir, fc.Package, fc.FileName)
	if filePath == "" {
		return fmt.Errorf("cannot find source file: %s/%s", fc.Package, fc.FileName)
	}

	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Print header
	fmt.Fprintf(w, "%s%s=== %s/%s ===%s\n", ColorBold, ColorWhite, fc.Package, fc.FileName, ColorReset)
	fmt.Fprintf(w, "%sCoverage: %.1f%% (%d/%d lines)%s\n\n", ColorWhite, fc.Coverage, fc.CoveredLines, fc.TotalLines, ColorReset)

	// Create a map for quick lookup of executable and executed lines
	executableMap := make(map[int]bool)
	for _, line := range fc.ExecutableLines {
		executableMap[line] = true
	}

	// Read and display file with coverage highlighting
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		lineText := scanner.Text()

		// Determine line color
		var lineColor string
		if !executableMap[lineNum] {
			// Non-executable line (comments, blank lines, etc.)
			lineColor = ColorGray
		} else if _, executed := fc.ExecutedLines[lineNum]; executed {
			// Executed line
			lineColor = ColorGreen
		} else {
			// Executable but not executed
			lineColor = ColorRed
		}

		// Print line with color
		fmt.Fprintf(w, "%s%4d%s %s%s%s\n",
			ColorWhite, lineNum, ColorReset,
			lineColor, lineText, ColorReset)

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Print uncovered lines summary
	uncovered := fc.GetUncoveredLines()
	if len(uncovered) > 0 {
		fmt.Fprintf(w, "\n%sUncovered lines:%s", ColorRed, ColorReset)
		for i, line := range uncovered {
			if i%10 == 0 {
				fmt.Fprintf(w, "\n  ")
			}
			fmt.Fprintf(w, "%d ", line)
		}
		fmt.Fprintln(w)
	}

	return nil
}

// findSourceFile attempts to locate the source file in the filesystem
func findSourceFile(rootDir, pkgPath, fileName string) string {
	// Try different possible locations
	candidates := []string{
		// Direct path
		filepath.Join(pkgPath, fileName),
		// Under examples
		filepath.Join(rootDir, "examples", pkgPath, fileName),
		// Under gnovm/stdlibs
		filepath.Join(rootDir, "gnovm", "stdlibs", pkgPath, fileName),
		// Just the package path from root
		filepath.Join(rootDir, pkgPath, fileName),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// GetFilesMatchingPattern returns all files matching the given pattern
func (t *Tracker) GetFilesMatchingPattern(pattern string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var files []string
	seen := make(map[string]bool)

	for pkg, pkgFiles := range t.executableLines {
		for fileName := range pkgFiles {
			fullName := fmt.Sprintf("%s/%s", pkg, fileName)
			if seen[fullName] {
				continue
			}

			if matched, _ := filepath.Match(pattern, fileName); matched {
				files = append(files, fullName)
				seen[fullName] = true
			} else if matched, _ := filepath.Match(pattern, filepath.Base(fileName)); matched {
				files = append(files, fullName)
				seen[fullName] = true
			}
		}
	}

	sort.Strings(files)
	return files
}
