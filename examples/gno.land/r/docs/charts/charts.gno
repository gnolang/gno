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
# Samcrew Package Documentation

## Available Packages

### 1. [Piechart](/p/samcrew/piechart) - SVG Pie Charts

Generate pie charts with legends as SVG markup for gnoweb rendering.

**Key Features:**
- SVG-based rendering with crisp graphics
- Customizable colors and labels
- Automatic legend generation with color swatches
- Mathematical precision for arc calculations

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

**Live Example:**

` + pieChartExample + `

---

*For complete API documentation, visit the individual package pages.*
`
}
