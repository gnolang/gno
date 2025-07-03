package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ANSI color codes for terminal output
const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorWhite = "\033[37m"
	ColorBold  = "\033[1m"
)

// Check if terminal supports colors
var supportsColor = true

func init() {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		supportsColor = false
		return
	}

	// Check if TERM environment variable indicates color support
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		supportsColor = false
	}
}

// ShowCoverage displays coverage visualization for files matching the pattern
func ShowCoverage(tracker *Tracker, pattern string, rootDir string) error {
	coverageData := tracker.GetCoverageData()

	// Convert glob pattern to regex
	regexPattern := globToRegex(pattern)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	found := false
	for filename, data := range coverageData {
		// Check if filename matches the pattern
		if !regex.MatchString(filepath.Base(filename)) {
			continue
		}

		found = true
		if err := visualizeFile(filename, data, rootDir); err != nil {
			return fmt.Errorf("failed to visualize %s: %w", filename, err)
		}
	}

	if !found {
		return fmt.Errorf("no files found matching pattern: %s", pattern)
	}

	return nil
}

// visualizeFile displays coverage visualization for a single file
func visualizeFile(filename string, data *CoverageData, rootDir string) error {
	// Try to find the actual file on disk
	filePath := findFileOnDisk(filename, rootDir)
	if filePath == "" {
		return fmt.Errorf("could not find file on disk: %s", filename)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fmt.Printf("\n%sCoverage visualization for: %s%s\n", ColorBold, filename, ColorReset)
	fmt.Printf("Total Lines: %d, Covered: %d, Coverage: %.2f%%\n",
		data.TotalLines, data.CoveredLines, data.CoverageRatio)

	if supportsColor {
		fmt.Printf("Legend: %sGreen%s = executed, %sRed%s = executable but not executed, %sWhite%s = non-executable or non-instrumented\n\n",
			ColorGreen, ColorReset, ColorRed, ColorReset, ColorWhite, ColorReset)
	} else {
		fmt.Printf("Legend: ✓ = executed, ✗ = executable but not executed, space = non-executable or non-instrumented\n\n")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 1
	uncoveredCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		color := getLineColor(lineNum, data)
		indicator := getLineIndicator(lineNum, data)

		// Count uncovered executable lines
		if indicator == "✗" {
			uncoveredCount++
		}

		// Format line number with padding
		lineNumStr := fmt.Sprintf("%4d", lineNum)

		if supportsColor {
			// Print line with color coding
			fmt.Printf("%s%s%s %s%s%s\n",
				ColorBold, lineNumStr, ColorReset,
				color, line, ColorReset)
		} else {
			// Print line with text indicator
			fmt.Printf("%s%s%s %s %s\n",
				ColorBold, lineNumStr, ColorReset,
				indicator, line)
		}

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if uncoveredCount > 0 {
		if supportsColor {
			fmt.Printf("\n%sUncovered executable lines: %d%s\n", ColorRed, uncoveredCount, ColorReset)
		} else {
			fmt.Printf("\nUncovered executable lines: %d\n", uncoveredCount)
		}
	}

	return nil
}

// getLineColor returns the appropriate color for a line based on coverage
func getLineColor(lineNum int, data *CoverageData) string {
	if !supportsColor {
		// Return empty string if colors are not supported
		return ""
	}

	if data == nil {
		return ColorWhite
	}

	if count, exists := data.LineData[lineNum]; exists {
		if count > 0 {
			return ColorGreen // Executed line - green
		} else {
			return ColorRed // Executable but not executed - red
		}
	}
	return ColorWhite // Non-instrumented or non-executable line - white
}

// getLineIndicator returns a text indicator for line coverage when colors are not supported
func getLineIndicator(lineNum int, data *CoverageData) string {
	if count, exists := data.LineData[lineNum]; exists {
		if count > 0 {
			return "✓" // Executed line
		} else {
			return "✗" // Executable but not executed
		}
	}
	return " " // Non-instrumented or non-executable line
}

// globToRegex converts a glob pattern to a regex pattern
func globToRegex(pattern string) string {
	// Escape special regex characters
	pattern = regexp.QuoteMeta(pattern)

	// Convert glob wildcards to regex
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")

	return "^" + pattern + "$"
}

// findFileOnDisk attempts to find the actual file on disk
func findFileOnDisk(filename string, rootDir string) string {
	// If filename contains path separators, try to resolve it
	if strings.Contains(filename, "/") {
		// Remove the package path prefix to get just the filename
		parts := strings.Split(filename, "/")
		justFilename := parts[len(parts)-1]

		// Try to find the file in examples directory first
		examplesPath := filepath.Join(rootDir, "examples", filename)
		if _, err := os.Stat(examplesPath); err == nil {
			return examplesPath
		}

		// Search recursively in the examples directory
		if found := findFileRecursively(filepath.Join(rootDir, "examples"), justFilename); found != "" {
			return found
		}
	}

	// Try different possible paths
	possiblePaths := []string{
		filepath.Join(rootDir, filename),
		filepath.Join(rootDir, "examples", filename),
		filepath.Join(rootDir, "gnovm", "stdlibs", filename),
		filepath.Join(rootDir, "gnovm", "tests", "stdlibs", filename),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Search recursively in the root directory
	return findFileRecursively(rootDir, filepath.Base(filename))
}

// findFileRecursively searches for a file recursively in a directory
func findFileRecursively(dir, filename string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively search subdirectories
			if found := findFileRecursively(path, filename); found != "" {
				return found
			}
		} else if entry.Name() == filename {
			return path
		}
	}

	return ""
}
