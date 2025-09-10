package piechart

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	title := "Test"
	slices := []PieSlice{
		{Value: 10, Color: "#00FF00", Label: "Green"},
		{Value: 20, Color: "#FFFF00", Label: "Yellow"},
		{Value: 30, Color: "#FF0000", Label: "Red"},
	}

	result := Render(slices, title)

	// Check if the result contains the expected image SVG structure
	if !strings.Contains(result, "data:image/svg+xml;base64,") {
		t.Errorf("Expected result to contain data:image/svg+xml;base64, got %s", result)
	}

	// Check if the result contains the title
	if !strings.Contains(result, title) {
		t.Errorf("SVG does not contain the title %q", title)
	}

	// Check if the result contains non-empty slices
	emptySlices := []PieSlice{}
	emptyResult := Render(emptySlices, title)
	expectedErr := "\npiechart fails: no data provided"
	if emptyResult != expectedErr {
		t.Errorf("Expected exact error for empty slices: %q, got %q", expectedErr, emptyResult)
	}
}
