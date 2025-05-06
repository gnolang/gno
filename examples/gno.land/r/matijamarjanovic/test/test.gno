package test

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/matijamarjanovic/charts"
)

var (
	testData = []float64{400.0, 1400.0, 200.0, 1600.0, 300.0, 1500.0, 250.0, 1450.0, 350.0, 1550.0, 280.0, 1420.0, 180.0, 1580.0, 320.0, 1480.0, 220.0, 1520.0, 380.0, 1460.0,
		150.0, 1550.0, 270.0, 1430.0, 310.0, 1590.0, 230.0, 1470.0, 360.0, 1510.0, 190.0, 1490.0, 260.0, 1540.0, 340.0, 1440.0, 210.0, 1570.0, 370.0, 1410.0,
		170.0, 1530.0, 290.0, 1450.0, 330.0, 1560.0, 240.0, 1480.0, 390.0, 1520.0, 160.0, 1460.0, 280.0, 1580.0, 320.0, 1420.0, 200.0, 1500.0, 350.0, 1470.0,
		180.0, 1540.0, 250.0, 1430.0, 360.0, 1570.0, 220.0, 1490.0, 310.0, 1510.0, 140.0, 1440.0, 270.0, 1550.0, 380.0, 1480.0, 230.0, 1520.0, 340.0, 1460.0,
		190.0, 1560.0, 260.0, 1420.0, 370.0, 1530.0, 210.0, 1470.0, 330.0, 1580.0, 150.0, 1450.0, 280.0, 1540.0, 390.0, 1410.0, 240.0, 1500.0, 320.0, 1480.0,
		170.0, 1550.0}
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
	output += "Here's the raw data with 100 values:\n"
	output += charts.GenerateColumnChart(testData, len(testData), "Raw Data", "Data Points", "Values") 
	output += "\n\n"

	output += "Here's the same data normalized to 12 columns:\n"
	output += charts.GenerateColumnChart(testData, 12, "Normalized Data", "Groups", "Values")
	output += "\n\n"

	output += "## Sample Column Chart\n\n"
	output += "Here's the raw data with 100 values:\n"
	output += charts.GenerateLineChart(testData, len(testData), 9, "Raw Data", "Data Points", "Values") 
	output += "\n\n"

	output += "Here's the same data normalized to 12 columns:\n"
	output += charts.GenerateLineChart(testData, 12, 9, "Normalized Data", "Groups", "Values")
	output += "\n\n"

	output += "## Sample Bar Chart\n\n"
	output += "Here's a horizontal bar chart example:\n\n"
	values := []float64{1081.25, 580, 182.5, 1081.5}
	labels := []string{"Apples", "Bananas", "Cherries", "Dates"}
	output += charts.GenerateBarChart(values, labels, 50, "Fruit Quantities")
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
