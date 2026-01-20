package lint

type Mode string

const (
	ModeDefault  Mode = "default"
	ModeStrict   Mode = "strict"
	ModeWarnOnly Mode = "warn-only"
)

// XXX: Add gnolint.toml config file support for project-level lint configuration.
type Config struct {
	Mode    Mode
	Disable map[string]bool // Rules to skip (e.g., {"AVL001": true})
}

func DefaultConfig() *Config {
	return &Config{
		Mode:    ModeDefault,
		Disable: make(map[string]bool),
	}
}

func (c *Config) IsRuleEnabled(ruleID string) bool {
	return !c.Disable[ruleID]
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
