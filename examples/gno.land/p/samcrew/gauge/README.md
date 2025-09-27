# `gauge` - SVG progress bars

Generate progress bars and gauges as SVG images with customizable styling and labels.

## Usage

```go
// Basic gauge with default config
gauge := gauge.Render(75, 100, "Progress", "#4caf50", gauge.DefaultConfig)

// Custom configuration
config := gauge.Config{
    PercentOnly:  true, // Show only percentage
    Width:        400,  // Custom width in pixels
    CanvasHeight: 40,   // Custom height in pixels
    FontSize:     18,   // Larger font size
    PaddingH:     10,   // More horizontal padding
}
gauge := gauge.Render(33, 50, "Loading", "#2196f3", config)

// Different styles
progress := gauge.Render(8, 10, "Health", "#f44336", gauge.DefaultConfig)
```

## API

```go
type Config struct {
    PercentOnly  bool // Show only percentage vs "value / total · percentage"
    Width        int  // Gauge width in pixels
    CanvasHeight int  // Height of the gauge in pixels
    FontSize     int  // Font size of the text in pixels
    PaddingH     int  // Horizontal padding (for the text) in pixels
}

var DefaultConfig = Config{false, 300, 30, 16, 6}

// `value`: Current value (must be ≤ total)
// `total`: Maximum value (must be > 0)
// `label`: Text label displayed on the left
// `color`: Fill color (hex format, e.g., "#4caf50")
// `config`: Configuration options
// Returns: SVG string as markdown image
func Render(value int, total int, label string, color string, config Config) string
```

**Output formats:**
- Default: `"Progress 75 / 100 · 75%"`
- PercentOnly: `"Progress 75%"`
