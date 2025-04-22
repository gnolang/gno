package coverage

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm"
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
		"testing.MarkLine(\"test.go\", 4)",  // if x > 0
		"testing.MarkLine(\"test.go\", 5)",  // return 1
		"testing.MarkLine(\"test.go\", 8)",  // return 0
	}

	t.Log("instrumented\n", string(instrumented))
	for _, exp := range expected {
		if !strings.Contains(string(instrumented), exp) {
			t.Errorf("expected instrumentation not found: %s", exp)
		}
	}
}

func TestInstrumentPackage(t *testing.T) {
	pkg := &gnovm.MemPackage{
		Files: []*gnovm.MemFile{
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
	if !strings.Contains(mainFile, "testing.MarkLine(\"main.gno\", 4)") {
		t.Error("Expected instrumentation in main.gno")
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
	pkg := &gnovm.MemPackage{
		Files: []*gnovm.MemFile{
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
