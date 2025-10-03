# `gauge` - SVG progress bars

Generate progress bars and gauges as SVG images with customizable styling and labels.

## Usage

```go
// Basic gauge with default config
gauge := gauge.Render(75, 100, "Progress", "#4caf50", gauge.DefaultConfig)

// Custom configuration
config := gauge.Config{
    PercentOnly:  true,
    Width:        400,
    CanvasHeight: 40,
    FontSize:     18,
    PaddingH:     10,
}
gauge := gauge.Render(33, 50, "Loading", "#2196f3", config)

// Different styles
progress := gauge.Render(8, 10, "Health", "#f44336", gauge.DefaultConfig)
```

## API Reference

```go
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
```

## Live Example

- [/r/docs/charts:gauge](/r/docs/charts:gauge)
