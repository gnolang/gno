// Package mdalert provides support for creating Markdown alerts.
//
// It defines supported alert types and helper functions that can be
// called to generate Markdown for different alert types.
//
// The different alert types are documented in the Markdown docs realm:
// https://gno.land/r/docs/markdown#alerts
package mdalert

import (
	"strings"

	"gno.land/p/nt/ufmt"
)

// Types of alerts.
const (
	TypeCaution Type = "CAUTION"
	TypeInfo         = "INFO"
	TypeNote         = "NOTE"
	TypeSuccess      = "SUCCESS"
	TypeTip          = "TIP"
	TypeWarning      = "WARNING"
)

type (
	// Type defines a type for the alert types.
	Type string

	// Alert defines a type for alerts.
	Alert struct {
		// Type defines the type of alert.
		Type Type

		// Title contains an optional title for the alert.
		Title string

		// Message contains alerts's message.
		Message string
	}
)

// String returns the alert as a Markdown string.
func (a Alert) String() string {
	alertType := string(a.Type)
	msg := strings.TrimSpace(a.Message)
	if msg == "" || alertType == "" {
		return ""
	}

	// Write alert header
	var b strings.Builder
	header := ufmt.Sprintf("> [!%s] %s", alertType, a.Title)
	b.WriteString(strings.TrimSpace(header) + "\n")

	// Write alert message
	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		b.WriteString("> " + line + "\n")
	}
	return b.String()
}

// New creates a new alert.
func New(t Type, title, msg string) Alert {
	return Alert{
		Type:    t,
		Title:   title,
		Message: msg,
	}
}

// Caution returns an alert Markdown of type caution.
func Caution(title, msg string) string {
	return New(TypeCaution, title, msg).String()
}

// Cautionf returns an alert Markdown of type caution with a formatted message.
func Cautionf(title, format string, a ...any) string {
	return New(TypeCaution, title, ufmt.Sprintf(format, a...)).String()
}

// Info returns an alert Markdown of type info.
func Info(title, msg string) string {
	return New(TypeInfo, title, msg).String()
}

// Infof returns an alert Markdown of type info with a formatted message.
func Infof(title, format string, a ...any) string {
	return New(TypeInfo, title, ufmt.Sprintf(format, a...)).String()
}

// Note returns an alert Markdown of type note.
func Note(title, msg string) string {
	return New(TypeNote, title, msg).String()
}

// Notef returns an alert Markdown of type note with a formatted message.
func Notef(title, format string, a ...any) string {
	return New(TypeNote, title, ufmt.Sprintf(format, a...)).String()
}

// Success returns an alert Markdown of type success.
func Success(title, msg string) string {
	return New(TypeSuccess, title, msg).String()
}

// Notef returns an alert Markdown of type success with a formatted message.
func Successf(title, format string, a ...any) string {
	return New(TypeSuccess, title, ufmt.Sprintf(format, a...)).String()
}

// Tip returns an alert Markdown of type tip.
func Tip(title, msg string) string {
	return New(TypeTip, title, msg).String()
}

// Tipf returns an alert Markdown of type tip with a formatted message.
func Tipf(title, format string, a ...any) string {
	return New(TypeTip, title, ufmt.Sprintf(format, a...)).String()
}

// Warning returns an alert Markdown of type warning.
func Warning(title, msg string) string {
	return New(TypeWarning, title, msg).String()
}

// Warningf returns an alert Markdown of type warning with a formatted message.
func Warningf(title, format string, a ...any) string {
	return New(TypeWarning, title, ufmt.Sprintf(format, a...)).String()
}
