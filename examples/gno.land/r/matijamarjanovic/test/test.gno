package test

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/matijamarjanovic/charts"
)

var (
	testData = []float64{1.0, 4.0, 2.0, 6.0, 3.0, 5.0, 2.5, 4.5, 3.5, 5.5, 2.8, 4.2, 1.8, 5.8, 3.2, 4.8, 2.2, 5.2, 3.8, 4.6,
		1.5, 5.5, 2.7, 4.3, 3.1, 5.9, 2.3, 4.7, 3.6, 5.1, 1.9, 4.9, 2.6, 5.4, 3.4, 4.4, 2.1, 5.7, 3.7, 4.1,
		1.7, 5.3, 2.9, 4.5, 3.3, 5.6, 2.4, 4.8, 3.9, 5.2, 1.6, 4.6, 2.8, 5.8, 3.2, 4.2, 2.0, 5.0, 3.5, 4.7,
		1.8, 5.4, 2.5, 4.3, 3.6, 5.7, 2.2, 4.9, 3.1, 5.1, 1.4, 4.4, 2.7, 5.5, 3.8, 4.8, 2.3, 5.2, 3.4, 4.6,
		1.9, 5.6, 2.6, 4.2, 3.7, 5.3, 2.1, 4.7, 3.3, 5.8, 1.5, 4.5, 2.8, 5.4, 3.9, 4.1, 2.4, 5.0, 3.2, 4.8,
		1.7, 5.5}
)

func Render(path string) string {
	switch path {
	case "":
		return renderHome()
	default:
		return "404 Not Found"
	}
}

func renderHome() string {
	var output string

	output += "# ASCII Chart Demo\n\n"
	output += "## Sample Column Chart\n\n"
	output += "Here's a chart with 100 values normalized to 4 columns:\n\n"
	output += charts.GenerateColumnChart(testData, 4, "Normalized Data", "Groups", "Values")
	output += "\n\n"

	output += "Here's the same data without normalization:\n\n"
	output += charts.GenerateColumnChart(testData, 100, "Raw Data", "Data Points", "Values")
	output += "\n\n"

	output += "## Test Data\n\n"
	output += "Raw values: "
	for i, v := range testData {
		if i > 0 {
			output += ", "
		}
		output += ufmt.Sprintf("%f", v)
	}

	return output
}
