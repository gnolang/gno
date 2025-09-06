# `gauge` - SVG progress bars

Generate progress bars and gauges as SVG images with customizable styling and labels.

## Usage

```go
// Basic gauge with default config
gauge := gauge.Render(75, 100, "Progress", "#4caf50", nil) // nil for &gauge.DefaultConfig

// Custom configuration
config := &gauge.Config{
    PercentOnly: true,  // Show only percentage
    Width:       400,   // Custom width in pixels
}
gauge := gauge.Render(33, 50, "Loading", "#2196f3", config)

// Different styles
progress := gauge.Render(8, 10, "Health", "#f44336", nil)
```

## API

```go
type Config struct {
    PercentOnly bool // Show only percentage vs "value / total · percentage"
    Width       int  // Gauge width in pixels
}

var DefaultConfig = Config{false, 300}

// `value`: Current value (must be ≤ total)
// `total`: Maximum value (must be > 0)
// `label`: Text label displayed on the left
// `color`: Fill color (hex format, e.g., "#4caf50")
// `config`: Configuration options
// Returns: SVG string as markdown image
func Render(value int, total int, label string, color string, config *Config) string
```

**Output formats:**
- Default: `"Progress 75 / 100 · 75%"`
- PercentOnly: `"Progress 75%"`
