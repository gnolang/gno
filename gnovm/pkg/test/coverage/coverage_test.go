package coverage

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCoverageInstrumenter_EmptyFile(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "empty.go")

	// empty file
	code := `package main`
	_, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	coverage := tracker.GetCoverage("empty.go")
	if len(coverage) != 0 {
		t.Errorf("coverage data found in empty file: %v", coverage)
	}
}

func TestCoverageInstrumenter_Visit(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// test function declaration
	code := `
package main

func testFunc() int {
	return 42
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// check if markLine call is added before the return statement
	expected := "testing.MarkLine(\"test.go\", 4)"
	if !strings.Contains(string(instrumented), expected) {
		t.Errorf("markLine call not added before the return statement")
	}
}

func TestCoverageInstrumenter_ElseIf(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test else if chain
	code := `
package main

func testFunc(x int) int {
	if x > 10 {
		return 1
	} else if x > 5 {
		return 2
	} else {
		return 3
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	// Check if all branches are instrumented
	expectedPatterns := []string{
		"testing.MarkLine(\"test.go\", 4)", // function entry
		"testing.MarkLine(\"test.go\", 5)", // if block
		"testing.MarkLine(\"test.go\", 7)", // else if block
		"testing.MarkLine(\"test.go\", 9)", // else block
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(instrumentedStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestCoverageInstrumenter_AnonymousFunction(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test anonymous function
	code := `
package main

func testFunc() {
	fn := func() {
		println("anonymous")
	}
	fn()
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	// Check if both function and anonymous function are instrumented
	expectedPatterns := []string{
		"testing.MarkLine(\"test.go\", 4)", // main function entry
		"testing.MarkLine(\"test.go\", 5)", // anonymous function entry
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(instrumentedStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestCoverageInstrumenter_DeferStatement(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test defer statement
	code := `
package main

func testFunc() {
	defer func() {
		println("cleanup")
	}()
	println("main")
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	coverage := tracker.GetCoverage("test.go")
	t.Logf("Coverage data: %+v", coverage)

	expectedPatterns := []string{
		"testing.MarkLine(\"test.go\", 4)", // function entry
		"testing.MarkLine(\"test.go\", 5)", // inside defer
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(instrumentedStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestCoverageInstrumenter_BranchStatements(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test break and continue statements
	code := `
package main

func testFunc() {
	for i := 0; i < 10; i++ {
		if i == 5 {
			break
		}
		if i%2 == 0 {
			continue
		}
		println(i)
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	// Break and continue should be registered as executable lines
	coverage := tracker.GetCoverage("test.go")
	t.Logf("Coverage data: %+v", coverage)

	// Should instrument switch entry AND each case
	expectedPatterns := []string{
		"testing.MarkLine(\"test.go\", 4)", // function entry
		"testing.MarkLine(\"test.go\", 5)", // loop entry
		"testing.MarkLine(\"test.go\", 6)", // inside i == 5 cond
		"testing.MarkLine(\"test.go\", 9)", // i%2 == 0 cond
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(instrumentedStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestCoverageInstrumenter_SwitchWithEntry(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test switch statement with entry instrumentation
	code := `
package main

func testFunc(x int) {
	switch x {
	case 1:
		println("one")
	case 2:
		println("two")
	default:
		println("other")
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	// Should instrument switch entry AND each case
	expectedPatterns := []string{
		"testing.MarkLine(\"test.go\", 4)",  // function entry
		"testing.MarkLine(\"test.go\", 6)",  // case 1
		"testing.MarkLine(\"test.go\", 8)",  // case 2
		"testing.MarkLine(\"test.go\", 10)", // default case
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(instrumentedStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestCoverageInstrumenter_Visit_EmptyFile(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "empty.gno")

	// test empty file
	code := `package main`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// check if markLine call is not added
	instrumentedStr := string(instrumented)
	if strings.Contains(instrumentedStr, "markLine") {
		t.Error("markLine call added to empty file")
	}
}

func TestCoverageInstrumenter_Visit_Comments(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "comments.gno")

	// test file with only comments
	code := `package main

// comment 1
// comment 2
// comment 3`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// check if markLine call is not added
	instrumentedStr := string(instrumented)
	if strings.Contains(instrumentedStr, "markLine") {
		t.Error("markLine call added to comments file")
	}
}

func TestCoverageInstrumenter_ControlFlow(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// test code with if statement and return statement
	code := `
package main

func testFunc(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// check if markLine call is added before the return statement
	expected := []string{
		"testing.MarkLine(\"test.go\", 4)", // function entry
		"testing.MarkLine(\"test.go\", 5)", // if block
		// Note: return statements are no longer instrumented separately
		// They are covered by the block instrumentation
	}

	t.Log("instrumented\n", string(instrumented))
	for _, exp := range expected {
		if !strings.Contains(string(instrumented), exp) {
			t.Errorf("expected instrumentation not found: %s", exp)
		}
	}
}

func TestInstrumentPackage(t *testing.T) {
	pkg := &std.MemPackage{
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: `package main

func main() {
	return 42
}`,
			},
			{
				Name: "main_test.gno",
				Body: `package main

func TestMain(t *testing.T) {
	t.Log("test")
}`,
			},
			{
				Name: "utils.gno",
				Body: `package main

func helper() int {
	return 0
}`,
			},
			{
				Name: "utils_test.gno",
				Body: `package main

func TestHelper(t *testing.T) {
	t.Log("test")
}`,
			},
		},
	}

	err := InstrumentPackage(pkg)
	if err != nil {
		t.Fatalf("Failed to instrument package: %v", err)
	}

	// Check instrumentation in main.gno
	mainFile := pkg.Files[0].Body
	if !strings.Contains(mainFile, "testing.MarkLine(\"main.gno\", 3)") {
		t.Error("Expected instrumentation in main.gno")
		t.Log("mainFile:\n", mainFile)
	}

	// Check no instrumentation in main_test.gno
	mainTestFile := pkg.Files[1].Body
	if strings.Contains(mainTestFile, "testing.MarkLine") {
		t.Error("Unexpected instrumentation in main_test.gno")
	}

	// Check instrumentation in utils.gno
	utilsFile := pkg.Files[2].Body
	t.Log("utilsFile\n", utilsFile)
	if !strings.Contains(utilsFile, "testing.MarkLine(\"utils.gno\", 3)") {
		t.Error("Expected instrumentation in utils.gno")
	}

	// Check no instrumentation in utils_test.gno
	utilsTestFile := pkg.Files[3].Body
	t.Log("utilsTestFile\n", utilsTestFile)
	if strings.Contains(utilsTestFile, "testing.MarkLine") {
		t.Error("Unexpected instrumentation in utils_test.gno")
	}
}

func TestInstrumentPackage_Error(t *testing.T) {
	pkg := &std.MemPackage{
		Files: []*std.MemFile{
			{
				Name: "invalid.gno",
				Body: `package main

func testFunc() int {
	return 42 // not closed`,
			},
		},
	}

	err := InstrumentPackage(pkg)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCoverageInstrumenter_Import(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		filename string
		want     string
	}{
		{
			name: "no imports",
			code: `package main

func test() int {
	return 42
}`,
			filename: "test.gno",
			want:     `import "testing"`,
		},
		{
			name: "with other imports",
			code: `package main

import "fmt"

func test() int {
	return 42
}`,
			filename: "test.gno",
			want:     `"testing"`, // combined import
		},
		{
			name: "already has testing import",
			code: `package main

import "testing"

func test() int {
	return 42
}`,
			filename: "test.gno",
			want:     `import "testing"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instrumenter := NewCoverageInstrumenter(NewCoverageTracker(), tt.filename)
			instrumented, err := instrumenter.InstrumentFile([]byte(tt.code))
			if err != nil {
				t.Fatalf("Failed to instrument code: %v", err)
			}

			t.Log("instrumented\n", string(instrumented))

			// Check if testing import is present
			if !strings.Contains(string(instrumented), tt.want) {
				t.Errorf("Expected testing import not found in instrumented code:\n%s", string(instrumented))
			}

			// Check if testing import is not duplicated
			count := strings.Count(string(instrumented), tt.want)
			if count > 1 {
				t.Errorf("Testing import is duplicated %d times in instrumented code:\n%s", count, string(instrumented))
			}
		})
	}
}

func TestInstrumentFileWithMultilineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkFor []string
	}{
		{
			name: "multiline comment between functions",
			input: `package test

func foo() {
	println("before comment")
}

/* Helper methods */

func bar() {
	println("after comment")
}`,
			wantErr: false,
			checkFor: []string{
				"/* Helper methods */",
				"testing.MarkLine",
				"func foo()",
				"func bar()",
			},
		},
		{
			name: "std.Emit with multiline arguments",
			input: `package test

import "std"

func mint() {
	std.Emit(
		"MintEvent",
		"to", "address",
		"tokenId", "123",
	)
}`,
			wantErr: false,
			checkFor: []string{
				"std.Emit(",
				"\"MintEvent\"",
				"testing.MarkLine",
			},
		},
		{
			name: "return statement in if block",
			input: `package test

func check() error {
	val, err := getValue()
	if err != nil {
		return err
	}
	return nil
}`,
			wantErr: false,
			checkFor: []string{
				"testing.MarkLine",
				"return err",
				"return nil",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewCoverageTracker()
			instrumenter := NewCoverageInstrumenter(tracker, "test.gno")

			result, err := instrumenter.InstrumentFile([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("InstrumentFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				resultStr := string(result)
				for _, check := range tt.checkFor {
					if !strings.Contains(resultStr, check) {
						t.Errorf("InstrumentFile() result missing %q\nGot:\n%s", check, resultStr)
					}
				}

				// Ensure the result is valid Go syntax by checking basic structure
				if strings.Contains(resultStr, ", )") {
					t.Errorf("InstrumentFile() produced invalid syntax with ', )'\nGot:\n%s", resultStr)
				}
			}
		})
	}
}

func TestCoverageTracker_MarkLine(t *testing.T) {
	tracker := NewCoverageTracker()

	// Test marking lines
	tracker.MarkLine("test.gno", 10)
	tracker.MarkLine("test.gno", 15)
	tracker.MarkLine("test.gno", 10) // Mark same line again

	coverage := tracker.GetCoverage("test.gno")
	if len(coverage) != 2 {
		t.Errorf("Expected 2 unique lines, got %d", len(coverage))
	}

	if coverage[10] != 2 {
		t.Errorf("Expected line 10 to be executed 2 times, got %d", coverage[10])
	}

	if coverage[15] != 1 {
		t.Errorf("Expected line 15 to be executed 1 time, got %d", coverage[15])
	}
}

func TestCoverageTracker_GetCoverageData(t *testing.T) {
	tracker := NewCoverageTracker()

	// Register executable lines
	tracker.RegisterExecutableLine("test.gno", 10)
	tracker.RegisterExecutableLine("test.gno", 15)
	tracker.RegisterExecutableLine("test.gno", 20)

	// Mark some lines as executed
	tracker.MarkLine("test.gno", 10)
	tracker.MarkLine("test.gno", 15)

	coverageData := tracker.GetCoverageData()

	if len(coverageData) != 1 {
		t.Errorf("Expected 1 file, got %d", len(coverageData))
	}

	data := coverageData["test.gno"]
	if data.TotalLines != 3 {
		t.Errorf("Expected 3 total lines, got %d", data.TotalLines)
	}

	if data.CoveredLines != 2 {
		t.Errorf("Expected 2 covered lines, got %d", data.CoveredLines)
	}

	expectedCoverage := float64(2) / float64(3) * 100
	if data.CoverageRatio != expectedCoverage {
		t.Errorf("Expected coverage %.2f%%, got %.2f%%", expectedCoverage, data.CoverageRatio)
	}

	// Check line data
	if data.LineData[10] != 1 {
		t.Errorf("Expected line 10 to have count 1, got %d", data.LineData[10])
	}
	if data.LineData[20] != 0 {
		t.Errorf("Expected line 20 to have count 0, got %d", data.LineData[20])
	}
}

func TestCoverageTracker_PrintCoverage(t *testing.T) {
	tracker := NewCoverageTracker()

	// Setup some coverage data
	tracker.RegisterExecutableLine("file1.gno", 10)
	tracker.RegisterExecutableLine("file1.gno", 15)
	tracker.RegisterExecutableLine("file2.gno", 20)

	tracker.MarkLine("file1.gno", 10)
	tracker.MarkLine("file2.gno", 20)

	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.PrintCoverage()

	w.Close()
	os.Stdout = originalStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check if output contains expected information
	if !strings.Contains(output, "Coverage Report:") {
		t.Error("Expected 'Coverage Report:' in output")
	}
	if !strings.Contains(output, "file1.gno") {
		t.Error("Expected 'file1.gno' in output")
	}
	if !strings.Contains(output, "file2.gno") {
		t.Error("Expected 'file2.gno' in output")
	}
}

func TestCoverageTracker_Reset(t *testing.T) {
	tracker := NewCoverageTracker()

	// Add some data
	tracker.RegisterExecutableLine("test.gno", 10)
	tracker.MarkLine("test.gno", 10)

	// Verify data exists
	coverage := tracker.GetCoverage("test.gno")
	if len(coverage) == 0 {
		t.Error("Expected coverage data before reset")
	}

	// Reset
	tracker.Reset()

	// Verify data is cleared
	coverage = tracker.GetCoverage("test.gno")
	if len(coverage) != 0 {
		t.Error("Expected no coverage data after reset")
	}

	coverageData := tracker.GetCoverageData()
	if len(coverageData) != 0 {
		t.Error("Expected no coverage data after reset")
	}
}

func TestGetGlobalTracker(t *testing.T) {
	tracker := GetGlobalTracker()
	if tracker == nil {
		t.Error("Expected non-nil global tracker")
	}

	// Test that it's the same instance
	tracker2 := GetGlobalTracker()
	if tracker != tracker2 {
		t.Error("Expected same global tracker instance")
	}
}

func TestCoverageInstrumenter_ForLoop(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	code := `
package main

func testFunc() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// Check if for loop is instrumented
	expected := "testing.MarkLine(\"test.go\", 5)"
	if !strings.Contains(string(instrumented), expected) {
		t.Errorf("Expected for loop instrumentation not found")
	}
}

func TestCoverageInstrumenter_RangeLoop(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	code := `
package main

func testFunc() {
	items := []int{1, 2, 3}
	for _, item := range items {
		println(item)
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// Check if range loop is instrumented (실제 라인: 6)
	expected := "testing.MarkLine(\"test.go\", 6)"
	if !strings.Contains(string(instrumented), expected) {
		t.Errorf("Expected range loop instrumentation not found")
	}
}

func TestCoverageInstrumenter_SwitchStatement(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	code := `
package main

func testFunc(x int) {
	switch x {
	case 1:
		println("one")
	case 2:
		println("two")
	default:
		println("other")
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}
	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n" + instrumentedStr)

	// we instrument both switch entry and each case
	expectedSwitch := "testing.MarkLine(\"test.go\", 4)"
	expectedCase1 := "testing.MarkLine(\"test.go\", 6)"
	expectedCase2 := "testing.MarkLine(\"test.go\", 8)"
	expectedDefault := "testing.MarkLine(\"test.go\", 10)"

	if !strings.Contains(instrumentedStr, expectedSwitch) {
		t.Errorf("Expected switch entry instrumentation not found")
	}
	if !strings.Contains(instrumentedStr, expectedCase1) {
		t.Errorf("Expected case 1 instrumentation not found")
	}
	if !strings.Contains(instrumentedStr, expectedCase2) {
		t.Errorf("Expected case 2 instrumentation not found")
	}
	if !strings.Contains(instrumentedStr, expectedDefault) {
		t.Errorf("Expected default case instrumentation not found")
	}
}

// TODO: Currently gno does not support select statement
func TestCoverageInstrumenter_SelectStatement(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	code := `
package main

func testFunc() {
	ch := make(chan int)
	select {
	case <-ch:
		println("received")
	default:
		println("default")
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}
	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n" + instrumentedStr)

	// we instrument both select entry and each case
	expectedSelect := "testing.MarkLine(\"test.go\", 4)"
	expectedCase := "testing.MarkLine(\"test.go\", 7)"
	expectedDefault := "testing.MarkLine(\"test.go\", 9)"

	if !strings.Contains(instrumentedStr, expectedSelect) {
		t.Errorf("Expected select entry instrumentation not found")
	}
	if !strings.Contains(instrumentedStr, expectedCase) {
		t.Errorf("Expected case instrumentation not found")
	}
	if !strings.Contains(instrumentedStr, expectedDefault) {
		t.Errorf("Expected default case instrumentation not found")
	}
}

// Test external instrumentation handling
func TestCoverageInstrumenter_ExternalInstrumentation(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test file with external instrumentation marker
	code := `
package main

func testFunc() {
	cross.Call()
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}
	// Should return original code without instrumentation
	if strings.Contains(string(instrumented), "testing.MarkLine") {
		t.Error("Expected no instrumentation for file with external instrumentation")
	}
	// But should still register executable lines
	coverage := tracker.GetCoverage("test.go")
	t.Logf("Coverage data: %+v", coverage)
	if len(coverage) != 0 {
		t.Error("Expected no coverage data for externally instrumented file")
	}
}

func TestCoverageInstrumenter_ComplexControlFlow(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	code := `
package main

func complexFunc(x, y int) int {
	if x > 0 {
		for i := 0; i < x; i++ {
			if i%2 == 0 {
				switch i {
				case 0:
					return 0
				case 2:
					return 2
				default:
					continue
				}
			}
		}
		return x
	} else {
		select {
		case <-make(chan int):
			return -1
		default:
			return y
		}
	}
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}
	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n" + instrumentedStr)

	// we instrument all control flow entries and branches
	expectedPatterns := []string{
		"testing.MarkLine(\"test.go\", 4)",  // function entry
		"testing.MarkLine(\"test.go\", 5)",  // if condition
		"testing.MarkLine(\"test.go\", 6)",  // for loop
		"testing.MarkLine(\"test.go\", 7)",  // nested if
		"testing.MarkLine(\"test.go\", 9)",  // case 0
		"testing.MarkLine(\"test.go\", 11)", // case 2
		"testing.MarkLine(\"test.go\", 13)", // default case
		"testing.MarkLine(\"test.go\", 19)", // else block
		"testing.MarkLine(\"test.go\", 21)", // case
		"testing.MarkLine(\"test.go\", 23)", // default
	}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(instrumentedStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestInstrumentPackage_WithErrors(t *testing.T) {
	// Test with nil package
	err := InstrumentPackage(nil)
	if err == nil {
		t.Error("Expected error for nil package")
	}

	// Test with invalid syntax
	pkg := &std.MemPackage{
		Files: []*std.MemFile{
			{
				Name: "invalid.gno",
				Body: `package main

func testFunc() {
	return 42 // missing closing brace`,
			},
		},
	}

	err = InstrumentPackage(pkg)
	if err == nil {
		t.Error("Expected error for invalid syntax")
	}
}

func TestCoverageInstrumenter_MultipleImports(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	code := `
package main

import (
	"fmt"
	"os"
)

func testFunc() {
	fmt.Println("test")
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	// Check that testing import is added correctly
	instrumentedStr := string(instrumented)
	if !strings.Contains(instrumentedStr, `"testing"`) {
		t.Error("Expected testing import to be added")
	}

	// Check that existing imports are preserved
	if !strings.Contains(instrumentedStr, `"fmt"`) {
		t.Error("Expected fmt import to be preserved")
	}
	if !strings.Contains(instrumentedStr, `"os"`) {
		t.Error("Expected os import to be preserved")
	}
}

func TestCoverageInstrumenter_AssignmentStatements(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test assignment statements with potential side effects
	code := `
package main

func testFunc() {
	x := getValue()
	y = 42
	z, err := getValueWithError()
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	// Check that function block is instrumented
	if !strings.Contains(instrumentedStr, "testing.MarkLine(\"test.go\", 4)") {
		t.Error("Expected function block instrumentation")
	}

	// Check that testing import is added
	if !strings.Contains(instrumentedStr, "import \"testing\"") {
		t.Error("Expected testing import to be added")
	}

	// Assignment statements should be registered as executable lines
	coverageData := tracker.GetCoverageData()
	if data, exists := coverageData["test.go"]; exists {
		t.Logf("Total executable lines: %d", data.TotalLines)
		if data.TotalLines == 0 {
			t.Error("Expected some executable lines to be registered")
		}
	}
}

func TestCoverageInstrumenter_ExpressionStatements(t *testing.T) {
	tracker := NewCoverageTracker()
	instrumenter := NewCoverageInstrumenter(tracker, "test.go")

	// Test expression statements with side effects
	code := `
package main

func testFunc() {
	println("hello")
	someFunc()
	obj.method()
}
`
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("instrumenting failed: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Instrumented code:\n", instrumentedStr)

	// Check that function block is instrumented
	if !strings.Contains(instrumentedStr, "testing.MarkLine(\"test.go\", 4)") {
		t.Error("Expected function block instrumentation")
	}

	// Expression statements should be registered as executable lines
	coverageData := tracker.GetCoverageData()
	if data, exists := coverageData["test.go"]; exists {
		t.Logf("Total executable lines: %d", data.TotalLines)
		if data.TotalLines == 0 {
			t.Error("Expected some executable lines to be registered")
		}
	}
}
