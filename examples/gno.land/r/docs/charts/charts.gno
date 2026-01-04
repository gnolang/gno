package charts

func Render(path string) string {
	// Handle subroutes for specific chart types
	switch path {
	case "piechart":
		return RenderPieChart()
	case "gauge":
		return RenderGauge()
	default:
		return RenderMainDocs()
	}
}

// RenderMainDocs renders the main charts documentation page
func RenderMainDocs() string {
	// Generate live examples
	pieChartExample := RenderPieChartExample()
	gaugeExample := RenderGaugeExample()

	return `
# Chart Packages Documentation

## Available Packages

- **[Piechart](/p/samcrew/piechart)** - SVG Pie Charts
  - [View All Examples](/r/docs/charts:piechart) - Detailed documentation with multiple examples
- **[Gauge](/p/samcrew/gauge)** - SVG Gauge Charts  
  - [View All Examples](/r/docs/charts:gauge) - Detailed documentation with multiple examples

## Examples

### [Piechart](/p/samcrew/piechart) - SVG Pie Charts
Generate pie charts with label as SVG images .

` + pieChartExample + `

[→ View all Piechart examples](/r/docs/charts:piechart)

### [Gauge](/p/samcrew/gauge) - SVG Gauge Charts  
Generate progress bars and gauges as SVG images with customizable styling and labels.

` + gaugeExample + `

[→ View all Gauge examples](/r/docs/charts:gauge)

---

*For complete API documentation, visit the individual package pages.*
`
}
