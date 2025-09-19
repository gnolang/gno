# `piechart` - SVG pie charts 

Generate pie charts with legends as SVG markup for gnoweb rendering.

## Usage

```go
type PieSlice struct {
	Value float64 // Value - "15"
	Color string // Hex color - "#ffffff"
	Label string // Value's name - "Mobile"
}
```

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

## Example

[/r/samcrew/daodemo/custom_condition:members](/r/samcrew/daodemo/custom_condition:members)