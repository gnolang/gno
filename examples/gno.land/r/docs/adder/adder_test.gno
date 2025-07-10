package adder

import (
	"testing"
)

func TestRenderAndAdd(t *testing.T) {
	// Initial Render output
	output := Render("")
	expected := `# Add Example

Current Number: 0

Last Updated: Never

[Increase Number](/r/docs/adder$help&func=Add&n=42)
`
	if output != expected {
		t.Errorf("Initial Render failed, got:\n%s", output)
	}

	// Call Add with a value of 10
	Add(cross, 10)

	// Call Add again with a value of -5
	Add(cross, -5)

	// Render after two Add calls
	finalOutput := Render("")

	// Initial Render output
	output = Render("")
	expected = `# Add Example

Current Number: 5

Last Updated: 2009-02-13 23:31:30

[Increase Number](/r/docs/adder$help&func=Add&n=42)
`
	if output != expected {
		t.Errorf("Final Render failed, got:\n%s\nexpected:\n%s", output, finalOutput)
	}
}
