# `piechart` - SVG pie charts 

Generate pie charts with legends as SVG markup for gnoweb rendering.

## Usage

```go
slices := []piechart.PieSlice{
    {Value: 30, Color: "#ff6b6b", Label: "Frontend"},
    {Value: 25, Color: "#4ecdc4", Label: "Backend"},
    {Value: 20, Color: "#45b7d1", Label: "DevOps"},
    {Value: 15, Color: "#96ceb4", Label: "Mobile"},
    {Value: 10, Color: "#ffeaa7", Label: "Other"},
}

// With title
titledChart := piechart.Render(slices, "Team Distribution")

// Without title  
untitledChart := piechart.Render(slices, "")
```

## API Reference

```go
type PieSlice struct {
    Value float64 // Numeric value for the slice
    Color string  // Hex color code (e.g., "#ff6b6b")
    Label string  // Display label for the slice
}

// slices: Array of PieSlice structs containing the data
// title: Chart title (empty string for no title)
// Returns: SVG markup as a string
func Render(slices []PieSlice, title string) string
```

## Live Example

- [/r/docs/charts:piechart](/r/docs/charts:piechart)
- [/r/samcrew/daodemo/custom_condition:members](/r/samcrew/daodemo/custom_condition:members)