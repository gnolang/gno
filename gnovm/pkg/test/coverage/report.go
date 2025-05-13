package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// CoverageReport define the structure of the coverage report
type CoverageReport struct {
	Files map[string]FileCoverage `json:"files"`
}

// FileCoverage define the structure of the file coverage
type FileCoverage struct {
	Lines    map[int]int `json:"lines"`
	Total    int         `json:"total"`
	Covered  int         `json:"covered"`
	Coverage float64     `json:"coverage"`
}

// GenerateReport generate the coverage report
func GenerateReport(tracker *CoverageTracker, outputFile string) error {
	report := CoverageReport{
		Files: make(map[string]FileCoverage),
	}

	// collect the coverage data for all files
	for filename, lines := range tracker.data {
		fileCoverage := FileCoverage{
			Lines: make(map[int]int),
		}

		// calculate the total and covered lines for the file
		total := 0
		covered := 0
		for line, count := range lines {
			fileCoverage.Lines[line] = count
			total++
			if count > 0 {
				covered++
			}
		}

		fileCoverage.Total = total
		fileCoverage.Covered = covered
		if total > 0 {
			fileCoverage.Coverage = float64(covered) / float64(total) * 100
		}

		report.Files[filename] = fileCoverage
	}

	// convert the report to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON conversion failed: %w", err)
	}

	// if the output file is specified, save it to a file
	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonData, 0o644); err != nil {
			return fmt.Errorf("failed to save file: %w", err)
		}
	} else {
		// if the output file is not specified, print it to the standard output
		fmt.Println(string(jsonData))
	}

	return nil
}

// PrintReport print the coverage report in a human-readable format
func PrintReport(tracker *CoverageTracker, w io.Writer) error {
	// sort the filenames
	filenames := make([]string, 0, len(tracker.data))
	for filename := range tracker.data {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// print the coverage information for each file
	for _, filename := range filenames {
		lines := tracker.data[filename]
		total := len(lines)
		covered := 0
		for _, count := range lines {
			if count > 0 {
				covered++
			}
		}

		coverage := float64(covered) / float64(total) * 100
		fmt.Fprintf(w, "%s: %.1f%% (%d/%d)\n",
			filepath.Base(filename),
			coverage,
			covered,
			total,
		)
	}

	return nil
}
