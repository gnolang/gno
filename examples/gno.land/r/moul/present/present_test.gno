package present

import (
	"strings"
	"testing"
)

func TestPresentation(t *testing.T) {
	// Test Set
	result := Set(
		"test1",
		"Test Presentation",
		"Test Event",
		"Test Author",
		"2024-01-01",
		"Content\n## Slide 1\n## Slide 2",
	)
	assertString(t, "presentation saved successfully", result)

	// Test invalid date
	result = Set(
		"test2",
		"Test Presentation 2",
		"Test Event 2",
		"Test Author 2",
		"invalid-date",
		"Content",
	)
	assertString(t, "400: invalid date format (expected: YYYY-MM-DD)", result)

	// Test Render list
	output := Render("")
	assertContains(t, output, "Test Presentation")
	assertContains(t, output, "Test Event")
	assertContains(t, output, "Test Author")
	assertContains(t, output, "2024-01-01")
	assertContains(t, output, "3") // number of slides (intro + 2)

	// Test Render view
	output = Render("view?view=test1")
	assertContains(t, output, "Test Presentation")
	assertContains(t, output, "Test Event")
	assertContains(t, output, "Test Author")
	assertContains(t, output, "2024-01-01")
	assertContains(t, output, "Content")
	assertContains(t, output, "## Slide 1")
	assertContains(t, output, "## Slide 2")

	// Test Render slides (intro)
	output = Render("slides?slides=test1&page=1")
	assertContains(t, output, "Test Presentation")
	assertContains(t, output, "Test Event")
	assertContains(t, output, "Test Author")
	assertContains(t, output, "2024-01-01")
	assertContains(t, output, "1/3")  // page indicator
	assertContains(t, output, "Next") // navigation

	// Test Render slides (middle)
	output = Render("slides?slides=test1&page=2")
	assertContains(t, output, "Slide 1")
	assertContains(t, output, "2/3")
	assertContains(t, output, "Prev")
	assertContains(t, output, "Next")

	// Test Render slides (last)
	output = Render("slides?slides=test1&page=3")
	assertContains(t, output, "Slide 2")
	assertContains(t, output, "3/3")
	assertContains(t, output, "Prev")

	// Test Delete
	result = Delete("test1")
	assertString(t, "presentation deleted successfully", result)

	// Test Delete non-existent
	result = Delete("nonexistent")
	assertString(t, "404: presentation not found", result)

	// Test view non-existent
	output = Render("view?view=test1")
	assertString(t, "404: presentation not found", output)
}

func TestMiami23(t *testing.T) {
	// Check if miami23 presentation was properly initialized
	output := Render("view?view=miami23")
	assertContains(t, output, "Portal Loop Demo (Miami 2023)")
	assertContains(t, output, "Miami Gno Tech Retreat 2023")
	assertContains(t, output, "Manfred Touron")
	assertContains(t, output, "2023-10-15")

	// Check number of slides
	assertContains(t, output, "Number of slides: 5") // intro + 4 sections

	// Check slide navigation
	output = Render("slides?slides=miami23&page=1")
	assertContains(t, output, "Portal Loop Demo")

	output = Render("slides?slides=miami23&page=2")
	assertContains(t, output, "Portal Loop")

	output = Render("slides?slides=miami23&page=3")
	assertContains(t, output, "Roadmap")
}

// Helper assertion functions
func assertString(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

func assertContains(t *testing.T, output, substring string) {
	t.Helper()
	if !strings.Contains(output, substring) {
		t.Errorf("Expected output to contain %q", substring)
	}
}
