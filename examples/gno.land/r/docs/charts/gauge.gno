package charts

import (
	"gno.land/p/samcrew/gauge"
)

// RenderGaugeExample creates a simple gauge bar example
func RenderGaugeExample() string {
	return gauge.Render(75, 100, "Progress", "#2ecc71", gauge.DefaultConfig)
}

// Render renders the gauge documentation page
func RenderGauge() string {
	// Multiple gauge examples
	basicExample := RenderGaugeExample()

	// Custom config example
	customConfig := gauge.Config{
		PercentOnly:  true,
		Width:        400,
		CanvasHeight: 35,
		FontSize:     14,
		PaddingH:     8,
	}
	customExample := gauge.Render(90, 100, "Loading", "#3498db", customConfig)

	// Different progress example
	progressExample := gauge.Render(33, 50, "Health", "#e74c3c", gauge.DefaultConfig)

	return `
# [Gauge](/p/samcrew/gauge) Documentation

Generate progress bars and gauges as SVG images with customizable styling and labels.

## Basic Usage

` + "```go" + `
import "gno.land/p/samcrew/gauge"

// Basic gauge with default config
gauge := gauge.Render(75, 100, "Progress", "#2ecc71", gauge.DefaultConfig)
` + "```" + `

` + basicExample + `

## Custom Configuration Example

` + "```go" + `
config := gauge.Config{
    PercentOnly:  true, // Show only percentage
    Width:        400,  // Custom width in pixels
    CanvasHeight: 35,   // Custom height in pixels
    FontSize:     14,   // Font size in pixels
    PaddingH:     8,    // Horizontal padding in pixels
}
gauge := gauge.Render(90, 100, "Loading", "#3498db", config)
` + "```" + `

` + customExample + `

## Different Progress Example

` + "```go" + `
healthGauge := gauge.Render(33, 50, "Health", "#e74c3c", gauge.DefaultConfig)
` + "```" + `

` + progressExample + `

## API Reference

` + "```go" + `
type Config struct {
    PercentOnly  bool // Only display the percentage on the right side
    Width        int  // Width of the gauge in pixels
    CanvasHeight int  // Height of the gauge in pixels
    FontSize     int  // Font size of the text in pixels
    PaddingH     int  // Horizontal padding (for the text) in pixels
}

var DefaultConfig = Config{
    PercentOnly:  false,
    Width:        300,
    CanvasHeight: 30,
    FontSize:     16,
    PaddingH:     6,
}

// value: Current value (must be ≤ total)
// total: Maximum value (must be > 0)
// label: Text label displayed on the left
// color: Fill color (hex format, e.g., "#4caf50")
// config: Configuration options
// Returns: SVG string as markdown image
func Render(value int, total int, label string, color string, config Config) string
` + "```" + `
---

[← Back to Charts](/r/docs/charts) | [View Package](/p/samcrew/gauge)
`
}
