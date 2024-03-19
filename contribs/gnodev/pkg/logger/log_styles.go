package logger

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// DefaultStyles returns the default styles.
func DefaultStyles() *log.Styles {
	style := log.DefaultStyles()
	style.Levels = map[log.Level]lipgloss.Style{
		log.DebugLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.DebugLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("63")),
		log.InfoLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.InfoLevel.String())).
			MaxWidth(1).
			Foreground(lipgloss.Color("12")),

		log.WarnLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.WarnLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("192")),
		log.ErrorLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.ErrorLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("204")),
		log.FatalLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.FatalLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("134")),
	}
	style.Keys = map[string]lipgloss.Style{
		"err": lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
		"error": lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
	}
	style.Values = map[string]lipgloss.Style{
		"err": lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
		"error": lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
	}

	return style
}
