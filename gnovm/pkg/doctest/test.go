package doctest

import (
	"fmt"
	"time"
)

type TestResult struct {
	Name     string
	Passed   bool
	Error    error
	Duration time.Duration
}

type TestSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Time    time.Duration
}

func RunTests(codeBlocks []codeBlock) TestSummary {
	summary := TestSummary{}
	startTime := time.Now()

	for _, cb := range codeBlocks {
		result := runTest(cb)
		printTestResult(result)
		updateSummary(&summary, result)
	}

	summary.Time = time.Since(startTime)
	printSummary(summary)
	return summary
}

func runTest(cd codeBlock) TestResult {
	start := time.Now()
	_, err := ExecuteCodeBlock(cd, GetStdlibsDir())
	duration := time.Since(start)

	return TestResult{
		// TODO add name field
		Passed:   err == nil,
		Error:    err,
		Duration: duration,
	}
}

func printTestResult(result TestResult) {
	status := "PASS"
	if !result.Passed {
		status = "FAIL"
	}
	fmt.Printf("--- %s: %s (%.2fs)\n", status, result.Name, result.Duration.Seconds())
	if !result.Passed {
		fmt.Printf("    %v\n", result.Error)
	}
}

func updateSummary(summary *TestSummary, result TestResult) {
	summary.Total++
	if result.Passed {
		summary.Passed++
	} else {
		summary.Failed++
	}
}

func printSummary(summary TestSummary) {
	fmt.Printf("\nTest Summary:\n")
	fmt.Printf("Total: %d, Passed: %d, Failed: %d, Skipped: %d\n",
		summary.Total, summary.Passed, summary.Failed, summary.Skipped)
	fmt.Printf("Time: %.2fs\n", summary.Time.Seconds())

	if summary.Failed > 0 {
		fmt.Printf("FAIL\n")
	} else {
		fmt.Printf("PASS\n")
	}
}
