package lint

type Mode string

const (
	ModeDefault  Mode = "default"
	ModeStrict   Mode = "strict"
	ModeWarnOnly Mode = "warn-only"
)

type Config struct {
	Mode   Mode
	Format string
	// Enable  []string // deferred
	// Disable []string // deferred
	// Nolint  NolintConfig // deferred - RequireReason bool
	// Rules   map[string]map[string]interface{} // deferred - per-rule config
}

func DefaultConfig() *Config {
	return &Config{
		Mode:   ModeDefault,
		Format: "text",
	}
}

// LoadConfig from gnolint.toml - deferred
// func LoadConfig(dir string) (*Config, error)

func (c *Config) IsRuleEnabled(ruleID string) bool {
	// When Enable/Disable lists are added, check them here
	return true
}

func (c *Config) EffectiveSeverity(s Severity) Severity {
	switch c.Mode {
	case ModeStrict:
		if s >= SeverityWarning {
			return SeverityError
		}
	case ModeWarnOnly:
		if s >= SeverityWarning {
			return SeverityWarning
		}
	}
	return s
}
