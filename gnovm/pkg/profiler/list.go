package profiler

import (
	"fmt"
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

	for _, sample := range p.Samples {
		// Check each location in the sample
		for _, loc := range sample.Location {
			// Check if this location's function matches our search
			// allows partial matching (e.g., "Sprintf" matches "gno.land/p/demo/ufmt.Sprintf")
			if strings.Contains(loc.Function, funcName) {
				// Get or create function data
				fd, exists := matchedFunctions[loc.Function]
				if !exists {
					fd = &functionData{
						funcName:    loc.Function,
						fileSamples: make(map[string]map[int]*lineStat),
					}
					matchedFunctions[loc.Function] = fd
				}

				file := loc.File
				line := loc.Line

				if file == "" || line == 0 {
					continue
				}

				// Skip test files when displaying function list
				if strings.HasSuffix(file, "_test.gno") {
					continue
				}

				if fd.fileSamples[file] == nil {
					fd.fileSamples[file] = make(map[int]*lineStat)
				}

				if fd.fileSamples[file][line] == nil {
					fd.fileSamples[file][line] = &lineStat{}
				}

				if len(sample.Value) > 0 {
					fd.fileSamples[file][line].count += sample.Value[0]
				}
				if len(sample.Value) > 1 {
					fd.fileSamples[file][line].cycles += sample.Value[1]
					fd.totalCycles += sample.Value[1]
				}
				// Add gas info if available
				if sample.GasUsed > 0 {
					fd.fileSamples[file][line].gas += sample.GasUsed
					fd.totalGas += sample.GasUsed
				}

				// Only count the first matching location in the stack
				break
			}
		}
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
				p.writeLineStatsOnly(w, lineStats, totalCycles)
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

	lines := strings.Split(source, "\n")

	// Find the function boundaries
	startLine, endLine := findFunctionBounds(lines, funcName)
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

	// Write lines with annotations - match go tool pprof format
	for i := showStart; i <= showEnd && i <= len(lines); i++ {
		line := ""
		if i <= len(lines) {
			line = lines[i-1]
		}

		if stat, exists := lineStats[i]; exists {
			// Line with profile data
			if stat.gas > 0 {
				fmt.Fprintf(w, "%10d %10d %4d:%s [gas: %d]\n",
					stat.cycles, stat.cycles, i, line, stat.gas)
			} else {
				fmt.Fprintf(w, "%10d %10d %4d:%s\n",
					stat.cycles, stat.cycles, i, line)
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
func (p *Profile) writeLineStatsOnly(w io.Writer, lineStats map[int]*lineStat, totalCycles int64) {
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
		}
	}

	// Try various common paths for Gno examples
	// This is necessary because the profiler only has the filename (e.g., "print.gno")
	// but needs to find the actual file location in the filesystem
	// Extract just the filename
	filename := filepath.Base(file)

	possiblePaths := []string{
		// Direct file path
		file,
		// Common Gno paths with filename
		filepath.Join("examples", "gno.land", "p", "demo", "ufmt", filename),
		filepath.Join("examples", "gno.land", "p", "demo", "int256", filename),
		filepath.Join("examples", "gno.land", "p", "demo", "avl", filename),
		filepath.Join("gnovm", "stdlibs", filename),
		filepath.Join("gnovm", "tests", "stdlibs", filename),
		// Try with package name in path (e.g., fmt/print.gno)
		// These paths were added to fix "<source not available>" for stdlib packages
		filepath.Join("gnovm", "tests", "stdlibs", "fmt", filename),
		filepath.Join("gnovm", "stdlibs", "fmt", filename),
		// Try to reconstruct path from package structure
		filepath.Join("examples", "gno.land", "p", "demo", filepath.Dir(file), filename),
		filepath.Join("gnovm", "stdlibs", filepath.Dir(file), filename),
		filepath.Join("gnovm", "tests", "stdlibs", filepath.Dir(file), filename),
	}

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
	}

	return "", fmt.Errorf("could not read source file: %s", file)
}

// findFunctionBounds tries to find the start and end lines of a function
func findFunctionBounds(lines []string, funcName string) (start, end int) {
	// Extract just the function name without package
	parts := strings.Split(funcName, ".")
	shortName := parts[len(parts)-1]

	inFunction := false
	braceCount := 0
	start = -1

	for i, line := range lines {
		// Look for function declaration
		if !inFunction && strings.Contains(line, "func ") && strings.Contains(line, shortName) {
			start = i + 1 // Convert to 1-based
			inFunction = true
			if strings.Contains(line, "{") {
				braceCount = 1
			}
			continue
		}

		if inFunction {
			// Count braces to find end
			for _, ch := range line {
				switch ch {
				case '{':
					braceCount++
				case '}':
					braceCount--
					if braceCount == 0 {
						return start, i + 1 // Convert to 1-based
					}
				default:
					continue
				}
			}
		}
	}

	// If we found start but not end, return to end of file
	if start != -1 {
		return start, len(lines)
	}

	return -1, -1
}
