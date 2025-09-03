package main

import (
	"fmt"
	"os"
	"strings"
)

// This program fixes the coverage profile by replacing incorrect file references
// It addresses the issue where coverage data is incorrectly attributed to testing.gno
// instead of the actual file being tested

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run fix_coverage.go <input_coverage_file> <output_coverage_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Read the coverage profile
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading coverage file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	fixedLines := make([]string, 0, len(lines))

	// First line is the mode declaration
	if len(lines) > 0 {
		fixedLines = append(fixedLines, lines[0])
	}

	// Process the coverage data lines
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		// Check if this is a line with "testing.gno:" prefix
		if strings.HasPrefix(line, "testing.gno:") {
			// Replace with the correct file
			line = strings.Replace(line, "testing.gno:", "cover.gno:", 1)
		}

		fixedLines = append(fixedLines, line)
	}

	// Write the fixed coverage profile
	err = os.WriteFile(outputFile, []byte(strings.Join(fixedLines, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error writing fixed coverage file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Coverage profile fixed and saved to %s\n", outputFile)
}