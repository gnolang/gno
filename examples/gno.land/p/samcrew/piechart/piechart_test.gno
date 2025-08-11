package piechart

import (
	"testing"
)

func TestGeneratePieChart(t *testing.T) {
	values := []float64{10, 20, 30}
	colors := []string{"#FF6384", "#36A2EB", "#FFCE56"}
	labels := []string{"Red", "Blue", "Yellow"}
	title := "Sample Pie Chart"

	result := GeneratePieChart(values, colors, labels, title)

	// Check if the result is not empty
	if result == "" {
		t.Error("Expected non-empty SVG string, got empty string")
	}

	// Check if one of the slices has not the same length than the others
	badValues := []float64{10, 20, 30, 40} // one extra value
	errMsg := "values, colors and labels slices must have the same length"

	if res := GeneratePieChart(badValues, colors, labels, title); res != errMsg {
		t.Errorf("Expected error message for mismatched lengths, got: %s", res)
	}
}
