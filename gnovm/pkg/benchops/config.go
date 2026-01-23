package benchops

// Config controls which measurements are enabled in the profiler.
type Config struct {
	EnableOps    bool
	EnableStore  bool
	EnableNative bool
}

// DefaultConfig returns a Config with all measurements enabled.
func DefaultConfig() Config {
	return Config{
		EnableOps:    true,
		EnableStore:  true,
		EnableNative: true,
	}
}
