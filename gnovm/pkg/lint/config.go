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
}

func DefaultConfig() *Config {
	return &Config{
		Mode:   ModeDefault,
		Format: "text",
	}
}

func (c *Config) IsRuleEnabled(ruleID string) bool {
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
