package test

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/matijamarjanovic/charts"
)

var (
	testData = []float64{1.0, 14.0, 2.0, 16.0, 3.0, 15.0, 2.5, 14.5, 3.5, 15.5, 2.8, 14.2, 1.8, 15.8, 3.2, 14.8, 2.2, 15.2, 3.8, 14.6,
		1.5, 15.5, 2.7, 14.3, 3.1, 15.9, 2.3, 14.7, 3.6, 15.1, 1.9, 14.9, 2.6, 15.4, 3.4, 14.4, 2.1, 15.7, 3.7, 14.1,
		1.7, 15.3, 2.9, 14.5, 3.3, 15.6, 2.4, 14.8, 3.9, 15.2, 1.6, 14.6, 2.8, 15.8, 3.2, 14.2, 2.0, 15.0, 3.5, 14.7,
		1.8, 15.4, 2.5, 14.3, 3.6, 15.7, 2.2, 14.9, 3.1, 15.1, 1.4, 14.4, 2.7, 15.5, 3.8, 14.8, 2.3, 15.2, 3.4, 14.6,
		1.9, 15.6, 2.6, 14.2, 3.7, 15.3, 2.1, 14.7, 3.3, 15.8, 1.5, 14.5, 2.8, 15.4, 3.9, 14.1, 2.4, 15.0, 3.2, 14.8,
		1.7, 15.5}
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
	output += charts.GenerateColumnChart(testData, 8, "Normalized Data", "Groups", "Values")
	output += "\n\n"

	output += "Here's the same data without normalization:\n\n"
	output += charts.GenerateColumnChart(testData, 100, "Raw Data", "Data Points", "Values")
	output += "\n\n"

	output += "## Sample Bar Chart\n\n"
	output += "Here's a horizontal bar chart example:\n\n"
	values := []float64{85, 60, 40, 75}
	labels := []string{"Apples", "Bananas", "Cherries", "Dates"}
	output += charts.GenerateBarChart(values, labels, 20, "Fruit Quantities")
	output += "\n\n"

	output += "## Test Data\n\n"
	output += "Raw values: "
	for i, v := range testData {
		if i > 0 {
			output += ", "
		}
		output += ufmt.Sprintf("%.2f", v)
	}

	return output
}
