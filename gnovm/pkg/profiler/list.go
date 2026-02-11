package profiler

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ROUTINE_SEPARATOR = "ROUTINE ========================"

// WriteFunctionList writes a line-by-line profile for a specific function,
// similar to 'go tool pprof list' command
func (p *Profile) WriteFunctionList(w io.Writer, funcName string, store Store) error {
	if p.Type == ProfileMemory {
		fmt.Fprintf(w, "Line-level listings are not available for memory profiles. Re-run with CPU or gas profiling plus -profile-line.\n")
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Group samples by function name that matches the search term
	type functionData struct {
		funcName    string
		fileSamples map[string]map[int]*lineStat // file -> line -> stats
		totalCycles int64
		totalGas    int64
	}

	matchedFunctions := make(map[string]*functionData)

	for name, data := range p.FunctionLines {
		if data == nil || !strings.Contains(name, funcName) {
			continue
		}
		fd := &functionData{
			funcName:    name,
			fileSamples: make(map[string]map[int]*lineStat),
			totalCycles: data.totalCycles,
			totalGas:    data.totalGas,
		}
		for file, lines := range data.fileSamples {
			fd.fileSamples[file] = make(map[int]*lineStat, len(lines))
			for line, stat := range lines {
				if stat == nil {
					continue
				}
				fd.fileSamples[file][line] = &lineStat{
					count:  stat.count,
					cycles: stat.cycles,
					gas:    stat.gas,
				}
			}
		}
		matchedFunctions[name] = fd
	}

	if len(matchedFunctions) == 0 {
		fmt.Fprintf(w, "No samples found for function: %s\n", funcName)
		return nil
	}

	// Sort functions by name for consistent output
	sortedFuncs := make([]string, 0, len(matchedFunctions))
	for fname := range matchedFunctions {
		sortedFuncs = append(sortedFuncs, fname)
	}
	sort.Strings(sortedFuncs)

	// Calculate total cycles and gas across all matched functions
	totalCycles := int64(0)
	totalGas := int64(0)
	for _, fd := range matchedFunctions {
		totalCycles += fd.totalCycles
		totalGas += fd.totalGas
	}

	// Print results for each matched function
	first := true
	for _, fname := range sortedFuncs {
		fd := matchedFunctions[fname]

		// Add separator between functions
		if !first {
			fmt.Fprintf(w, "\n")
		}
		first = false

		// Print results for each file in this function
		for file, lineStats := range fd.fileSamples {
			if err := p.writeFunctionFileList(w, fd.funcName, file, lineStats, totalCycles, totalGas, store); err != nil {
				// If we can't read the source file, still show the statistics
				fmt.Fprintf(w, "\nTotal: %d\n", totalCycles)
				fmt.Fprintf(w, "%s %s in %s\n", ROUTINE_SEPARATOR, fd.funcName, file)
				fileCycles := int64(0)
				for _, stat := range lineStats {
					fileCycles += stat.cycles
				}
				fmt.Fprintf(w, "%10d %10d (flat, cum) %.2f%% of Total\n", fileCycles, fileCycles,
					float64(fileCycles)/float64(totalCycles)*100)
				p.writeLineStatsOnly(w, lineStats)
			}
		}
	}

	return nil
}

// writeFunctionFileList writes the profile for a single file
func (p *Profile) writeFunctionFileList(w io.Writer, funcName, file string, lineStats map[int]*lineStat, totalCycles int64, totalGas int64, store Store) error {
	// Try to read the source file
	source, err := readSourceFile(file, store)
	if err != nil {
		return err
	}

	// Find the function boundaries
	startLine, endLine, err := findFunctionBounds(source, funcName)
	if err != nil {
		return err
	}

	lines := strings.Split(source, "\n")
	if startLine == -1 {
		// If we can't find the function, show all lines with samples
		startLine = 1
		endLine = len(lines)
	}

	// Calculate total cycles and gas for this file
	fileCycles := int64(0)
	fileGas := int64(0)
	for _, stat := range lineStats {
		fileCycles += stat.cycles
		fileGas += stat.gas
	}

	// Write header - enhanced with gas info
	fmt.Fprintf(w, "\nTotal: %d cycles", totalCycles)
	if totalGas > 0 {
		fmt.Fprintf(w, ", %d gas", totalGas)
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s %s in %s\n", ROUTINE_SEPARATOR, funcName, file)
	fmt.Fprintf(w, "%10d %10d (flat, cum) %.2f%% of Total", fileCycles, fileCycles,
		float64(fileCycles)/float64(totalCycles)*100)
	if fileGas > 0 {
		fmt.Fprintf(w, " | Gas: %d (%.2f%%)", fileGas, float64(fileGas)/float64(totalGas)*100)
	}
	fmt.Fprintf(w, "\n")

	// Show context lines before and after
	contextLines := 5
	showStart := max(1, startLine-contextLines)
	showEnd := min(len(lines), endLine+contextLines)

	cumulativeCycles := int64(0)

	// Write lines with annotations - match go tool pprof format
	for i := showStart; i <= showEnd && i <= len(lines); i++ {
		line := ""
		if i <= len(lines) {
			line = lines[i-1]
		}

		if stat, exists := lineStats[i]; exists {
			cumulativeCycles += stat.cycles
			// Line with profile data
			if stat.gas > 0 {
				fmt.Fprintf(w, "%10d %10d %4d:%s [gas: %d]\n",
					stat.cycles, cumulativeCycles, i, line, stat.gas)
			} else {
				fmt.Fprintf(w, "%10d %10d %4d:%s\n",
					stat.cycles, cumulativeCycles, i, line)
			}
		} else {
			// Lines without samples
			if i >= startLine && i <= endLine {
				// Inside function, show dots
				fmt.Fprintf(w, "         . %10s %4d:%s\n", ".", i, line)
			} else {
				// Context lines outside function
				fmt.Fprintf(w, "           %10s %4d:%s\n", "", i, line)
			}
		}
	}

	return nil
}

// writeLineStatsOnly writes just the statistics when source is not available
func (p *Profile) writeLineStatsOnly(w io.Writer, lineStats map[int]*lineStat) {
	// Sort lines
	lines := make([]int, 0, len(lineStats))
	for line := range lineStats {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	// Show line numbers with cycle counts
	for _, line := range lines {
		stat := lineStats[line]
		fmt.Fprintf(w, "%10d %10d %4d:<source not available>\n",
			stat.cycles, stat.cycles, line)
	}
}

var stdlibPaths = []string{"fmt", "std", "strings", "strconv", "math"}

// readSourceFile attempts to read a source file
func readSourceFile(file string, store Store) (string, error) {
	// First try to read as absolute path
	content, err := os.ReadFile(file)
	if err == nil {
		return string(content), nil
	}

	// Try relative to current directory
	content, err = os.ReadFile(filepath.Join(".", file))
	if err == nil {
		return string(content), nil
	}

	// Try to read from store if available
	if store != nil {
		// Extract package path and filename from the file path
		// This is a simplified approach - in practice you'd need better parsing
		parts := strings.Split(file, "/")
		if len(parts) >= 2 {
			pkgPath := strings.Join(parts[:len(parts)-1], "/")
			fileName := parts[len(parts)-1]

			if memFile := store.GetMemFile(pkgPath, fileName); memFile != nil {
				return memFile.Body, nil
			}
		} else {
			if !strings.HasSuffix(file, ".gno") {
				return "", fmt.Errorf("could not read source file: %s", file)
			}
			// Handle cases where we only have a filename without a package path.
			// This commonly occurs with standard library files where the profiler
			// receives just "print.gno" instead of "fmt/print.gno".
			// Try common stdlib paths
			for _, pkg := range stdlibPaths {
				if memFile := store.GetMemFile(pkg, file); memFile != nil {
					return memFile.Body, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not read source file: %s", file)
}

func findFunctionBounds(source string, funcName string) (start int, end int, err error) {
	fset := token.NewFileSet()

	// ParseMode = parser.ParseComments | parser.AllErrors
	file, parseErr := parser.ParseFile(fset, "", source, parser.ParseComments)
	if parseErr != nil {
		return -1, -1, parseErr
	}

	shortName := funcName
	// handle x.y.z → extract z
	if dot := strings.LastIndex(shortName, "."); dot >= 0 {
		shortName = shortName[dot+1:]
	}

	var foundStart, foundEnd int

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if decl.Name.Name != shortName {
			return true
		}

		// Found target function
		startPos := decl.Body.Pos()
		endPos := decl.Body.End()

		start = fset.Position(startPos).Line
		end = fset.Position(endPos).Line

		foundStart, foundEnd = start, end
		return false // stop traversal
	})

	if foundStart == 0 {
		return -1, -1, errors.New("function not found: " + funcName)
	}

	return foundStart, foundEnd, nil
}
