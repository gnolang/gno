package coverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// InteractiveViewer handles interactive coverage viewing
type InteractiveViewer struct {
	tracker     *Tracker
	rootDir     string
	files       []string
	currentIdx  int
	writer      io.Writer
	reader      io.Reader
	scanner     *bufio.Scanner
	currentLine int // Current line position in the file
	viewHeight  int // Number of lines to display at once
}

// NewInteractiveViewer creates a new interactive viewer
func NewInteractiveViewer(tracker *Tracker, rootDir string, files []string, r io.Reader, w io.Writer) *InteractiveViewer {
	return &InteractiveViewer{
		tracker:     tracker,
		rootDir:     rootDir,
		files:       files,
		currentIdx:  0,
		writer:      w,
		reader:      r,
		currentLine: 1,
		viewHeight:  30, // Default view height, can be adjusted based on terminal size
	}
}

// Start begins the interactive viewing session
func (iv *InteractiveViewer) Start() error {
	if len(iv.files) == 0 {
		return fmt.Errorf("no files to display")
	}

	// Initialize scanner for line-based input
	iv.scanner = bufio.NewScanner(iv.reader)

	// Clear screen and show initial file
	iv.clearScreen()
	if err := iv.displayCurrentFile(); err != nil {
		return err
	}

	// Main interaction loop
	for {
		// Show prompt
		fmt.Fprintf(iv.writer, "\n%s[File %d/%d, Line %d] n/p=files, :n=line, J/K=page, gg/G=top/end, h=help > %s",
			ColorBold, iv.currentIdx+1, len(iv.files), iv.currentLine, ColorReset)

		// Read command
		if !iv.scanner.Scan() {
			return nil // EOF
		}

		cmd := strings.TrimSpace(iv.scanner.Text())
		if cmd == "" {
			cmd = "n" // Default to next
		}

		// Process command
		switch cmd[0] {
		case 'n', 'j': // Next file
			if iv.currentIdx < len(iv.files)-1 {
				iv.currentIdx++
				iv.currentLine = 1 // Reset line position for new file
				iv.clearScreen()
				if err := iv.displayCurrentFile(); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(iv.writer, "%sAlready at last file%s\n", ColorRed, ColorReset)
			}

		case 'p', 'k': // Previous file
			if iv.currentIdx > 0 {
				iv.currentIdx--
				iv.currentLine = 1 // Reset line position for new file
				iv.clearScreen()
				if err := iv.displayCurrentFile(); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(iv.writer, "%sAlready at first file%s\n", ColorRed, ColorReset)
			}

		case ':': // Vim-style line navigation
			if len(cmd) > 1 {
				lineStr := cmd[1:]
				if lineNum, err := strconv.Atoi(lineStr); err == nil && lineNum > 0 {
					iv.currentLine = lineNum
					iv.clearScreen()
					if err := iv.displayCurrentFileAtLine(); err != nil {
						return err
					}
				} else {
					fmt.Fprintf(iv.writer, "%sInvalid line number: %s%s\n", ColorRed, lineStr, ColorReset)
				}
			} else {
				fmt.Fprintf(iv.writer, "%sUsage: :<line number>%s\n", ColorRed, ColorReset)
			}

		case 'J': // Page down
			iv.currentLine += iv.viewHeight
			iv.clearScreen()
			if err := iv.displayCurrentFileAtLine(); err != nil {
				return err
			}

		case 'K': // Page up
			iv.currentLine -= iv.viewHeight
			if iv.currentLine < 1 {
				iv.currentLine = 1
			}
			iv.clearScreen()
			if err := iv.displayCurrentFileAtLine(); err != nil {
				return err
			}

		case 'g': // Either goto file or beginning of file
			if cmd == "gg" {
				// Go to beginning of file
				iv.currentLine = 1
				iv.clearScreen()
				if err := iv.displayCurrentFileAtLine(); err != nil {
					return err
				}
			} else if strings.HasPrefix(cmd, "g ") {
				// Original goto file number functionality
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					if num, err := strconv.Atoi(parts[1]); err == nil && num >= 1 && num <= len(iv.files) {
						iv.currentIdx = num - 1
						iv.currentLine = 1 // Reset line position for new file
						iv.clearScreen()
						if err := iv.displayCurrentFile(); err != nil {
							return err
						}
					} else {
						fmt.Fprintf(iv.writer, "%sInvalid file number. Range: 1-%d%s\n",
							ColorRed, len(iv.files), ColorReset)
					}
				}
			} else {
				fmt.Fprintf(iv.writer, "%sUnknown command. Use 'gg' for beginning of file or 'g <number>' for file navigation%s\n", ColorRed, ColorReset)
			}

		case 'G': // Go to end of file
			// We'll need to get the total lines count for the current file
			report := iv.tracker.GenerateReport()
			currentFile := iv.files[iv.currentIdx]
			for _, fc := range report.Files {
				fullName := fmt.Sprintf("%s/%s", fc.Package, fc.FileName)
				if fullName == currentFile {
					// Set to a large number, it will be adjusted in displayCurrentFileAtLine
					iv.currentLine = 99999
					iv.clearScreen()
					if err := iv.displayCurrentFileAtLine(); err != nil {
						return err
					}
					break
				}
			}

		case 'l': // List all files
			iv.listFiles()

		case 'h', '?': // Help
			iv.showHelp()

		case 'q': // Quit
			return nil

		default:
			fmt.Fprintf(iv.writer, "%sUnknown command. Type 'h' for help.%s\n", ColorRed, ColorReset)
		}
	}
}

// clearScreen clears the terminal screen
func (iv *InteractiveViewer) clearScreen() {
	switch runtime.GOOS {
	case "windows":
		fmt.Fprint(iv.writer, "\033c")
	default:
		// ANSI escape sequence to clear screen and move cursor to top
		fmt.Fprint(iv.writer, "\033[2J\033[H")
	}
}

// displayCurrentFile shows the current file's coverage
func (iv *InteractiveViewer) displayCurrentFile() error {
	iv.currentLine = 1 // Reset to beginning when showing full file
	return iv.displayCurrentFileAtLine()
}

// displayCurrentFileAtLine shows the current file's coverage starting from a specific line
func (iv *InteractiveViewer) displayCurrentFileAtLine() error {
	// Display header with navigation info
	fmt.Fprintf(iv.writer, "%s%s=== Coverage Viewer ===%s\n",
		ColorBold, ColorWhite, ColorReset)

	// Get file coverage
	report := iv.tracker.GenerateReport()
	var fileCoverage *FileCoverage
	currentFile := iv.files[iv.currentIdx]

	for _, fc := range report.Files {
		fullName := fmt.Sprintf("%s/%s", fc.Package, fc.FileName)
		if fullName == currentFile {
			fileCoverage = fc
			break
		}
	}

	if fileCoverage == nil {
		return fmt.Errorf("coverage data not found for file: %s", currentFile)
	}

	// Display file info with current line indicator
	fmt.Fprintf(iv.writer, "%sFile: %s/%s (Line %d)%s\n",
		ColorWhite, fileCoverage.Package, fileCoverage.FileName, iv.currentLine, ColorReset)

	// Display the file coverage with line offset
	return iv.showFileWithLineOffset(fileCoverage)
}

// listFiles shows a list of all available files
func (iv *InteractiveViewer) listFiles() {
	fmt.Fprintf(iv.writer, "\n%s%sAvailable files:%s\n", ColorBold, ColorWhite, ColorReset)

	report := iv.tracker.GenerateReport()
	for i, fileName := range iv.files {
		// Find coverage info
		var coverage float64
		for _, fc := range report.Files {
			fullName := fmt.Sprintf("%s/%s", fc.Package, fc.FileName)
			if fullName == fileName {
				coverage = fc.Coverage
				break
			}
		}

		// Highlight current file
		if i == iv.currentIdx {
			fmt.Fprintf(iv.writer, "%s> %d. %s (%.1f%%)%s\n",
				ColorGreen, i+1, fileName, coverage, ColorReset)
		} else {
			fmt.Fprintf(iv.writer, "  %d. %s (%.1f%%)\n", i+1, fileName, coverage)
		}
	}
}

// showHelp displays help information
func (iv *InteractiveViewer) showHelp() {
	fmt.Fprintf(iv.writer, "\n%s%sCoverage Viewer Commands:%s\n",
		ColorBold, ColorWhite, ColorReset)

	helpText := []struct {
		cmd  string
		desc string
	}{
		{"n, j, <enter>", "Go to next file"},
		{"p, k", "Go to previous file"},
		{"g <number>", "Go to specific file by number"},
		{":<line>", "Jump to specific line in current file"},
		{"gg", "Go to beginning of current file"},
		{"G", "Go to end of current file"},
		{"J", "Page down (scroll down one screen)"},
		{"K", "Page up (scroll up one screen)"},
		{"l", "List all files with coverage"},
		{"h, ?", "Show this help"},
		{"q", "Quit viewer"},
	}

	for _, item := range helpText {
		fmt.Fprintf(iv.writer, "  %s%-15s%s - %s\n",
			ColorGreen, item.cmd, ColorReset, item.desc)
	}
}

// showFileWithLineOffset displays the file coverage starting from the current line
func (iv *InteractiveViewer) showFileWithLineOffset(fc *FileCoverage) error {
	// Try to find the file
	filePath := findSourceFile(iv.rootDir, fc.Package, fc.FileName)
	if filePath == "" {
		return fmt.Errorf("cannot find source file: %s/%s", fc.Package, fc.FileName)
	}

	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Print coverage summary
	fmt.Fprintf(iv.writer, "%sCoverage: %.1f%% (%d/%d lines)%s\n",
		ColorWhite, fc.Coverage, fc.CoveredLines, fc.TotalLines, ColorReset)
	fmt.Fprintf(iv.writer, "%s[Viewing from line %d, use :<n> to jump to line n]%s\n\n",
		ColorGray, iv.currentLine, ColorReset)

	// Create a map for quick lookup of executable and executed lines
	executableMap := make(map[int]bool)
	for _, line := range fc.ExecutableLines {
		executableMap[line] = true
	}

	// Read all lines first to determine total lines
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	totalLines := len(lines)

	// Adjust current line if it's beyond the file
	if iv.currentLine > totalLines {
		iv.currentLine = totalLines
	}
	if iv.currentLine < 1 {
		iv.currentLine = 1
	}

	// Calculate display range
	startLine := iv.currentLine - 1 // Convert to 0-based index
	if startLine < 0 {
		startLine = 0
	}
	if startLine >= totalLines {
		startLine = totalLines - 1
	}

	endLine := startLine + iv.viewHeight
	if endLine > totalLines {
		endLine = totalLines
	}

	// Display lines in range
	for i := startLine; i < endLine; i++ {
		lineNum := i + 1
		lineText := lines[i]

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

		// Highlight current line
		lineMarker := " "
		if lineNum == iv.currentLine {
			lineMarker = ">"
		}

		// Print line with color
		fmt.Fprintf(iv.writer, "%s%s%4d%s %s%s%s\n",
			ColorWhite, lineMarker, lineNum, ColorReset,
			lineColor, lineText, ColorReset)
	}

	// Show navigation info if there are more lines
	if endLine < totalLines {
		fmt.Fprintf(iv.writer, "\n%s... %d more lines (total: %d) ...%s\n",
			ColorGray, totalLines-endLine, totalLines, ColorReset)
	}

	return nil
}
