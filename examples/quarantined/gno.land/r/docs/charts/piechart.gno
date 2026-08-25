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

// Render renders the piechart documentation page
func RenderPieChart() string {
	// Multiple piechart examples
	basicExample := RenderPieChartExample()

	// Custom example with different data
	customData := []piechart.PieSlice{
		{Value: 40, Color: "#FF6B6B", Label: "Red Section"},
		{Value: 30, Color: "#4ECDC4", Label: "Teal Section"},
		{Value: 20, Color: "#45B7D1", Label: "Blue Section"},
		{Value: 10, Color: "#96CEB4", Label: "Green Section"},
	}
	customExample := piechart.Render(customData, "Custom Color Example")

	// Small dataset example
	smallData := []piechart.PieSlice{
		{Value: 70, Color: "#FF9F43", Label: "Majority"},
		{Value: 30, Color: "#686DE0", Label: "Minority"},
	}
	smallExample := piechart.Render(smallData, "Simple Split")

	return `
# [Piechart](/p/samcrew/piechart) Documentation

Generate pie charts with label as SVG images.

## Basic Usage

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

` + basicExample + `

---

## Custom Colors Example

` + "```go" + `
customData := []piechart.PieSlice{
    {Value: 40, Color: "#FF6B6B", Label: "Red Section"},
    {Value: 30, Color: "#4ECDC4", Label: "Teal Section"},
    {Value: 20, Color: "#45B7D1", Label: "Blue Section"},
    {Value: 10, Color: "#96CEB4", Label: "Green Section"},
}
chart := piechart.Render(customData, "Custom Color Example")
` + "```" + `

` + customExample + `

---

## Simple Split Example

` + "```go" + `
simpleData := []piechart.PieSlice{
    {Value: 70, Color: "#FF9F43", Label: "Majority"},
    {Value: 30, Color: "#686DE0", Label: "Minority"},
}
chart := piechart.Render(simpleData, "Simple Split")
` + "```" + `

` + smallExample + `

## API Reference

` + "```go" + `
type PieSlice struct {
    Value float64 // Numeric value for the slice
    Color string  // Hex color code (e.g., "#ff6b6b")
    Label string  // Display label for the slice
}

// slices: Array of PieSlice structs containing the data
// title: Chart title (empty string for no title)
// Returns: SVG markup as a string
func Render(slices []PieSlice, title string) string
` + "```" + `

---

[← Back to Charts](/r/docs/charts) | [View Package](/p/samcrew/piechart)
`
}
