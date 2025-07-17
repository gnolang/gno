package coverage

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

// InteractiveViewer handles interactive coverage viewing
type InteractiveViewer struct {
	tracker    *Tracker
	rootDir    string
	files      []string
	currentIdx int
	writer     io.Writer
	reader     io.Reader
	scanner    *bufio.Scanner
}

// NewInteractiveViewer creates a new interactive viewer
func NewInteractiveViewer(tracker *Tracker, rootDir string, files []string, r io.Reader, w io.Writer) *InteractiveViewer {
	return &InteractiveViewer{
		tracker:    tracker,
		rootDir:    rootDir,
		files:      files,
		currentIdx: 0,
		writer:     w,
		reader:     r,
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
		fmt.Fprintf(iv.writer, "\n%s[%d/%d] (n)ext, (p)rev, (g)oto, (l)ist, (h)elp, (q)uit > %s",
			ColorBold, iv.currentIdx+1, len(iv.files), ColorReset)

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
				iv.clearScreen()
				if err := iv.displayCurrentFile(); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(iv.writer, "%sAlready at first file%s\n", ColorRed, ColorReset)
			}

		case 'g': // Goto file number
			parts := strings.Fields(cmd)
			if len(parts) > 1 {
				if num, err := strconv.Atoi(parts[1]); err == nil && num >= 1 && num <= len(iv.files) {
					iv.currentIdx = num - 1
					iv.clearScreen()
					if err := iv.displayCurrentFile(); err != nil {
						return err
					}
				} else {
					fmt.Fprintf(iv.writer, "%sInvalid file number. Range: 1-%d%s\n",
						ColorRed, len(iv.files), ColorReset)
				}
			} else {
				fmt.Fprintf(iv.writer, "%sUsage: g <number>%s\n", ColorRed, ColorReset)
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

	// Display the file coverage
	return iv.tracker.showSingleFileCoverage(iv.rootDir, fileCoverage, iv.writer)
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
		{"l", "List all files with coverage"},
		{"h, ?", "Show this help"},
		{"q", "Quit viewer"},
	}

	for _, item := range helpText {
		fmt.Fprintf(iv.writer, "  %s%-15s%s - %s\n",
			ColorGreen, item.cmd, ColorReset, item.desc)
	}
}
