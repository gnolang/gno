package coverage

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCoverageTracker_InvariantValidation(t *testing.T) {
	tracker := NewTracker()

	// Set up test data that satisfies all invariants
	tracker.RegisterExecutableLine("test.gno", 10)
	tracker.RegisterExecutableLine("test.gno", 15)
	tracker.RegisterExecutableLine("test.gno", 20)

	tracker.MarkLine("test.gno", 10)
	tracker.MarkLine("test.gno", 15)

	// Validate that all invariants hold
	if err := tracker.ValidateInvariants(); err != nil {
		t.Errorf("Invariants should be satisfied, got error: %v", err)
	}

	// Test Invariant I3: Coverage ratio in [0, 100]
	coverageData := tracker.GetCoverageData()
	data := coverageData["test.gno"]
	if data.CoverageRatio < 0 || data.CoverageRatio > 100 {
		t.Errorf("Invariant I3 violated: coverage ratio %f not in [0, 100]", data.CoverageRatio)
	}
}

func TestCoverageTracker_InvariantViolation(t *testing.T) {
	tracker := NewTracker()

	// Manually create invalid state to test invariant checking
	tracker.data = map[string]map[int]int{
		"test.gno": {10: 1}, // Line 10 executed but not registered
	}
	tracker.allLines = map[string]map[int]bool{
		"test.gno": {15: true}, // Only line 15 registered
	}

	// This should violate Invariant I1
	if err := tracker.ValidateInvariants(); err == nil {
		t.Error("Expected invariant violation, but validation passed")
	}
}

func TestCrossIdentifierDetector(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "file with cross identifier",
			code: `package main
func test() {
	cross.Call()
}`,
			expected: true,
		},
		{
			name: "file without cross identifier",
			code: `package main
func test() {
	normal.Call()
}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewInstrumentationEngine(NewTracker(), "test.gno")
			content, err := engine.InstrumentFile([]byte(tt.code))
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			hasInstrumentation := strings.Contains(string(content), "testing.MarkLine")

			if tt.expected && hasInstrumentation {
				t.Error("Expected no instrumentation for externally instrumented file")
			}
			if !tt.expected && !hasInstrumentation {
				t.Error("Expected instrumentation for normal file")
			}
		})
	}
}

func TestDefaultBranchingStrategy(t *testing.T) {
	// Test function entry instrumentation
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")
	code := `package main
func testFunc() {
	return
}`
	content, err := engine.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("Failed to instrument: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "testing.MarkLine") {
		t.Error("Expected function entry instrumentation")
	}
}

func TestFunctionRule_R1(t *testing.T) {
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")

	tests := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "regular function",
			code: `package main
func testFunc() {
	return 42
}`,
			want: []string{"testing.MarkLine(\"test.gno\", 2)"},
		},
		{
			name: "anonymous function",
			code: `package main
func testFunc() {
	fn := func() {
		return 42
	}
	fn()
}`,
			want: []string{
				"testing.MarkLine(\"test.gno\", 2)", // main function
				"testing.MarkLine(\"test.gno\", 3)", // anonymous function
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := engine.InstrumentFile([]byte(tt.code))
			if err != nil {
				t.Fatalf("Failed to instrument: %v", err)
			}

			contentStr := string(content)
			for _, pattern := range tt.want {
				if !strings.Contains(contentStr, pattern) {
					t.Errorf("Expected pattern not found: %s\nGenerated:\n%s", pattern, contentStr)
				}
			}
		})
	}
}

func TestConditionalRule_R2_ElseIf(t *testing.T) {
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")

	code := `package main
func testFunc(x int) {
	if x > 10 {
		return 1
	} else if x > 5 {
		return 2
	} else {
		return 3
	}
}`

	content, err := engine.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("Failed to instrument: %v", err)
	}

	contentStr := string(content)
	t.Log("Instrumented code:\n", contentStr)

	// Rule R2: All branches should be instrumented independently
	expectedPatterns := []string{
		"testing.MarkLine(\"test.gno\", 2)", // function entry
		"testing.MarkLine(\"test.gno\", 3)", // if branch
		"testing.MarkLine(\"test.gno\", 5)", // else if branch
		"testing.MarkLine(\"test.gno\", 7)", // else branch
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestSwitchSelectRule_R4_EntryInstrumentation(t *testing.T) {
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")

	code := `package main
func testFunc(x int) {
	switch x {
	case 1:
		return 1
	case 2:
		return 2
	default:
		return 0
	}
}`

	content, err := engine.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("Failed to instrument: %v", err)
	}

	contentStr := string(content)
	t.Log("Instrumented code:\n", contentStr)

	// Rule R4: Both switch entry and each case should be instrumented
	expectedPatterns := []string{
		"testing.MarkLine(\"test.gno\", 2)", // function entry
		"testing.MarkLine(\"test.gno\", 4)", // case 1
		"testing.MarkLine(\"test.gno\", 6)", // case 2
		"testing.MarkLine(\"test.gno\", 8)", // default
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestStatementRule_IndividualInstrumentation(t *testing.T) {
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")

	code := `package main
func testFunc() {
	x := getValue()
	y = 42
	println("test")
	someFunc()
}`

	content, err := engine.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("Failed to instrument: %v", err)
	}

	contentStr := string(content)
	t.Log("Instrumented code:\n", contentStr)

	// Statement-level instrumentation: each assignment and expression should be instrumented
	expectedPatterns := []string{
		"testing.MarkLine(\"test.gno\", 2)", // function entry
		"testing.MarkLine(\"test.gno\", 3)", // x := getValue()
		"testing.MarkLine(\"test.gno\", 4)", // y = 42
		"testing.MarkLine(\"test.gno\", 5)", // println("test")
		"testing.MarkLine(\"test.gno\", 6)", // someFunc()
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("Expected pattern not found: %s", pattern)
		}
	}
}

func TestPrinciple_P1_Symmetry(t *testing.T) {
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")

	// Test that similar constructs are instrumented similarly
	tests := []struct {
		name string
		code string
	}{
		{
			name: "for loop",
			code: `package main
func test() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}`,
		},
		{
			name: "range loop",
			code: `package main
func test() {
	items := []int{1, 2, 3}
	for _, item := range items {
		println(item)
	}
}`,
		},
	}

	instrumentedCodes := make([]string, 0, len(tests))
	for _, tt := range tests {
		content, err := engine.InstrumentFile([]byte(tt.code))
		if err != nil {
			t.Fatalf("Failed to instrument %s: %v", tt.name, err)
		}
		instrumentedCodes = append(instrumentedCodes, string(content))
	}

	// Both should have similar instrumentation patterns (function + loop)
	for i, code := range instrumentedCodes {
		if !strings.Contains(code, "testing.MarkLine") {
			t.Errorf("Test %d should have instrumentation", i)
		}
		// Count instrumentation calls - should be similar structure
		count := strings.Count(code, "testing.MarkLine")
		if count < 2 { // At least function + loop
			t.Errorf("Test %d should have at least 2 instrumentation points, got %d", i, count)
		}
	}
}

func TestPrinciple_P2_MinimalIntrusion(t *testing.T) {
	engine := NewInstrumentationEngine(NewTracker(), "test.gno")

	originalCode := `package main

// Important comment
func testFunc() {
	x := 42 // Another comment
	return x
}`

	content, err := engine.InstrumentFile([]byte(originalCode))
	if err != nil {
		t.Fatalf("Failed to instrument: %v", err)
	}

	contentStr := string(content)

	// Comments should be preserved (Principle P2)
	if !strings.Contains(contentStr, "// Important comment") {
		t.Error("Important comment should be preserved")
	}
	if !strings.Contains(contentStr, "// Another comment") {
		t.Error("Inline comment should be preserved")
	}

	// Original structure should be preserved
	if !strings.Contains(contentStr, "x := 42") {
		t.Error("Original code structure should be preserved")
	}
}

func TestPrinciple_P3_Completeness(t *testing.T) {
	tracker := NewTracker()
	engine := NewInstrumentationEngine(tracker, "test.gno")

	code := `package main
func complexFunc(x int) int {
	if x > 0 {
		for i := 0; i < x; i++ {
			switch i {
			case 0:
				return 0
			default:
				continue
			}
		}
	}
	return -1
}`

	_, err := engine.InstrumentFile([]byte(code))
	if err != nil {
		t.Fatalf("Failed to instrument: %v", err)
	}

	// Principle P3: All executable paths should be tracked
	coverageData := tracker.GetCoverageData()
	data := coverageData["test.gno"]

	if data.TotalLines == 0 {
		t.Error("Expected some executable lines to be registered")
	}

	// Should cover: function, if, for, switch, case, default, return statements
	minExpectedLines := 7
	if data.TotalLines < minExpectedLines {
		t.Errorf("Expected at least %d executable lines, got %d", minExpectedLines, data.TotalLines)
	}
}

func TestInstrumentPackage(t *testing.T) {
	pkg := &std.MemPackage{
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: `package main

func main() {
	x := getValue()
	if x > 0 {
		println("positive")
	} else {
		println("non-positive")
	}
}`,
			},
			{
				Name: "utils.gno",
				Body: `package main

func getValue() int {
	return 42
}`,
			},
			{
				Name: "main_test.gno", // Should be skipped
				Body: `package main

func TestMain(t *testing.T) {
	main()
}`,
			},
		},
	}

	err := InstrumentPackage(pkg)
	if err != nil {
		t.Fatalf("Failed to instrument package: %v", err)
	}

	// Verify that non-test files are instrumented
	mainFile := pkg.Files[0].Body
	if !strings.Contains(mainFile, "testing.MarkLine") {
		t.Error("main.gno should be instrumented")
	}

	utilsFile := pkg.Files[1].Body
	if !strings.Contains(utilsFile, "testing.MarkLine") {
		t.Error("utils.gno should be instrumented")
	}

	// Verify that test files are not instrumented
	testFile := pkg.Files[2].Body
	if strings.Contains(testFile, "testing.MarkLine") {
		t.Error("test files should not be instrumented")
	}

	// Verify axiom compliance by checking invariants
	if err := GetGlobalTracker().ValidateInvariants(); err != nil {
		t.Errorf("Axiom system invariants violated: %v", err)
	}
}

func TestAxiomSystemIntegration(t *testing.T) {
	// Test that the entire system works together following the axiom system
	tracker := NewTracker()
	engine := NewInstrumentationEngine(tracker, "integration.gno")

	complexCode := `package main

import "fmt"

func complexExample(n int) string {
	// Assignment with side effect (A1: Executability)
	result := fmt.Sprintf("Processing %d", n)
	
	// Control flow branching (A3: Branching)
	if n < 0 {
		return "negative"
	} else if n == 0 {
		return "zero"
	} else {
		// Loop with multiple exit points
		for i := 0; i < n; i++ {
			switch i % 3 {
			case 0:
				if i == 6 {
					break
				}
				continue
			case 1:
				result += " odd-div-3"
			default:
				result += " even-div-3"
			}
		}
	}
	
	// Defer statement (R6)
	defer fmt.Println("Cleanup")
	
	return result
}`

	instrumented, err := engine.InstrumentFile([]byte(complexCode))
	if err != nil {
		t.Fatalf("Failed to instrument complex code: %v", err)
	}

	instrumentedStr := string(instrumented)
	t.Log("Complex instrumented code:\n", instrumentedStr)

	// Verify comprehensive instrumentation
	instrumentationCount := strings.Count(instrumentedStr, "testing.MarkLine")
	if instrumentationCount < 10 {
		t.Errorf("Expected comprehensive instrumentation, got only %d points", instrumentationCount)
	}

	// Verify that all axioms are satisfied
	if err := tracker.ValidateInvariants(); err != nil {
		t.Errorf("Axiom system invariants violated in complex example: %v", err)
	}

	// Verify original imports are preserved (P2: Minimal Intrusion)
	if !strings.Contains(instrumentedStr, `"fmt"`) {
		t.Error("Original imports should be preserved")
	}
}
