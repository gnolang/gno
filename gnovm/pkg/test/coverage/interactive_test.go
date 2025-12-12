package coverage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInteractiveViewerCommands(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	testPkg := "test/package"

	// Create package directory
	pkgDir := filepath.Join(tmpDir, "test", "package")
	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test files
	file1Content := `package test

func Hello() {
	println("Hello")
}
`
	file2Content := `package test

func World() {
	println("World")
}
`

	err = os.WriteFile(filepath.Join(pkgDir, "file1.gno"), []byte(file1Content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(pkgDir, "file2.gno"), []byte(file2Content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test tracker with sample data
	tracker := &Tracker{
		executableLines: make(map[string]map[string]map[int]bool),
		coverage:        make(map[string]map[string]map[int]int64),
		enabled:         true,
	}

	// Add test data
	file1 := "file1.gno"
	file2 := "file2.gno"

	tracker.executableLines[testPkg] = map[string]map[int]bool{
		file1: {3: true, 4: true},
		file2: {3: true, 4: true},
	}

	tracker.coverage[testPkg] = map[string]map[int]int64{
		file1: {3: 5, 4: 2},
		file2: {3: 1},
	}

	tests := []struct {
		name     string
		input    string
		files    []string
		wantErr  bool
		checkOut func(t *testing.T, output string)
	}{
		{
			name:  "quit command",
			input: "q\n",
			files: []string{"test/package/file1.gno", "test/package/file2.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Coverage Viewer") {
					t.Error("Expected output to contain 'Coverage Viewer'")
				}
			},
		},
		{
			name:  "help command",
			input: "h\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Coverage Viewer Commands") {
					t.Error("Expected help text")
				}
				if !strings.Contains(output, "Go to next file") {
					t.Error("Expected command descriptions")
				}
			},
		},
		{
			name:  "list command",
			input: "l\nq\n",
			files: []string{"test/package/file1.gno", "test/package/file2.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Available files:") {
					t.Error("Expected file list header")
				}
				if !strings.Contains(output, "file1.gno") {
					t.Error("Expected file1.gno in list")
				}
				if !strings.Contains(output, "file2.gno") {
					t.Error("Expected file2.gno in list")
				}
			},
		},
		{
			name:  "next command",
			input: "n\nq\n",
			files: []string{"test/package/file1.gno", "test/package/file2.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// Should show file navigation
				if !strings.Contains(output, "File 1/2") || !strings.Contains(output, "File 2/2") {
					t.Error("Expected navigation indicators")
				}
			},
		},
		{
			name:  "previous command at start",
			input: "p\nq\n",
			files: []string{"test/package/file1.gno", "test/package/file2.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Already at first file") {
					t.Error("Expected 'already at first file' message")
				}
			},
		},
		{
			name:  "goto command",
			input: "g 2\nq\n",
			files: []string{"test/package/file1.gno", "test/package/file2.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// After goto 2, should be at File 2/2
				if !strings.Contains(output, "File 2/2") {
					t.Error("Expected to be at file 2 after goto")
				}
			},
		},
		{
			name:  "goto invalid number",
			input: "g 5\nq\n",
			files: []string{"test/package/file1.gno", "test/package/file2.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Invalid file number") {
					t.Error("Expected invalid file number error")
				}
			},
		},
		{
			name:  "unknown command",
			input: "x\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Unknown command") {
					t.Error("Expected unknown command message")
				}
			},
		},
		{
			name:    "no files error",
			input:   "",
			files:   []string{},
			wantErr: true,
		},
		{
			name:  "line jump command",
			input: ":3\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// Should show that we're viewing from line 3
				if !strings.Contains(output, "Line 3") {
					t.Error("Expected to show line 3 indicator")
				}
			},
		},
		{
			name:  "invalid line jump",
			input: ":abc\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "Invalid line number") {
					t.Error("Expected invalid line number error")
				}
			},
		},
		{
			name:  "go to beginning (gg)",
			input: ":10\ngg\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// Should be at line 1 after gg
				if !strings.Contains(output, "Line 1") {
					t.Error("Expected to be at line 1 after gg")
				}
			},
		},
		{
			name:  "go to end (G)",
			input: "G\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// Should show we're at the end of the file
				if !strings.Contains(output, "Line") {
					t.Error("Expected to show line indicator")
				}
			},
		},
		{
			name:  "page down (J)",
			input: "J\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// Should have scrolled down
				if !strings.Contains(output, "Line") {
					t.Error("Expected to show line indicator after page down")
				}
			},
		},
		{
			name:  "page up (K)",
			input: ":50\nK\nq\n",
			files: []string{"test/package/file1.gno"},
			checkOut: func(t *testing.T, output string) {
				t.Helper()
				// Should have scrolled up
				if !strings.Contains(output, "Line") {
					t.Error("Expected to show line indicator after page up")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			viewer := NewInteractiveViewer(tracker, tmpDir, tt.files, input, output)

			// Start interactive session
			err := viewer.Start()
			if (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check output if provided
			if tt.checkOut != nil && err == nil {
				tt.checkOut(t, output.String())
			}
		})
	}
}

func TestClearScreen(t *testing.T) {
	viewer := &InteractiveViewer{
		writer: &bytes.Buffer{},
	}

	viewer.clearScreen()

	output := viewer.writer.(*bytes.Buffer).String()
	// Should contain ANSI escape sequence for clearing screen
	if !strings.Contains(output, "\033[2J\033[H") && !strings.Contains(output, "\033c") {
		t.Error("Expected ANSI escape sequence for clearing screen")
	}
}

func TestListFiles(t *testing.T) {
	tracker := &Tracker{
		executableLines: make(map[string]map[string]map[int]bool),
		coverage:        make(map[string]map[string]map[int]int64),
		enabled:         true,
	}

	// Add test data
	pkg := "test/package"
	file1 := "file1.gno"
	file2 := "file2.gno"

	tracker.executableLines[pkg] = map[string]map[int]bool{
		file1: {1: true, 2: true, 3: true, 4: true},
		file2: {1: true, 2: true},
	}

	tracker.coverage[pkg] = map[string]map[int]int64{
		file1: {1: 1, 2: 1, 3: 1}, // 3/4 = 75%
		file2: {1: 1},             // 1/2 = 50%
	}

	output := &bytes.Buffer{}
	viewer := &InteractiveViewer{
		tracker:    tracker,
		files:      []string{"test/package/file1.gno", "test/package/file2.gno"},
		currentIdx: 0,
		writer:     output,
	}

	viewer.listFiles()

	out := output.String()
	// Check that both files are listed with correct coverage
	if !strings.Contains(out, "file1.gno (75.0%)") {
		t.Error("Expected file1.gno with 75.0% coverage")
	}
	if !strings.Contains(out, "file2.gno (50.0%)") {
		t.Error("Expected file2.gno with 50.0% coverage")
	}
	// Check that current file is highlighted
	if !strings.Contains(out, "> 1. test/package/file1.gno") {
		t.Error("Expected current file to be highlighted")
	}
}

func TestShowHelp(t *testing.T) {
	output := &bytes.Buffer{}
	viewer := &InteractiveViewer{
		writer: output,
	}

	viewer.showHelp()

	out := output.String()
	// Check for key commands
	expectedCommands := []string{
		"n, j, <enter>",
		"p, k",
		"g <number>",
		"l",
		"h, ?",
		"q",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(out, cmd) {
			t.Errorf("Expected help to contain command: %s", cmd)
		}
	}
}
