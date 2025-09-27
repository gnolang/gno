package charts

import (
	"gno.land/p/samcrew/piechart"
)

// RenderPieChartExample creates a simple data distribution pie chart example
func RenderPieChartExample() string {
	data := []piechart.PieSlice{
		{Value: 35, Color: "#e74c3c", Label: "Category A"},
		{Value: 25, Color: "#3498db", Label: "Category B"},
		{Value: 20, Color: "#2ecc71", Label: "Category C"},
		{Value: 15, Color: "#f39c12", Label: "Category D"},
		{Value: 5, Color: "#9b59b6", Label: "Category E"},
	}
	return piechart.Render(data, "Data Distribution Example")
}

func Render(path string) string {
	// Generate live examples
	pieChartExample := RenderPieChartExample()

	return `
# Chart Packages Documentation

## Available Packages

- [Charts](/p/matijamarjanovic/charts) - SVG Chart Library
- [Gauge](/p/samcrew/gauge) - SVG Gauge Charts
- [Piechart](/p/samcrew/piechart) - SVG Pie Charts

## Live Examples
### [Piechart](/p/samcrew/piechart) - SVG Pie Charts

**Usage:**
` + "```go" + `
import "gno.land/p/samcrew/piechart"

data := []piechart.PieSlice{
    {Value: 35, Color: "#e74c3c", Label: "Category A"},
    {Value: 25, Color: "#3498db", Label: "Category B"},
    {Value: 20, Color: "#2ecc71", Label: "Category C"},
    {Value: 15, Color: "#f39c12", Label: "Category D"},
    {Value: 5, Color: "#9b59b6", Label: "Category E"},
}
chart := piechart.Render(data, "Data Distribution Example")
` + "```" + `

` + pieChartExample + `

---

*For complete API documentation, visit the individual package pages.*
`
}
