// Alert provides GitHub-style alert/callout components for markdown.
//
// These render as blockquotes with special markers that GitHub and
// compatible markdown renderers display as colored callout boxes.
//
// Example usage:
//
//	doc := mictofun.New().
//	    H1("Documentation").
//	    AddRaw(mictofun.Note("This is important information.")).
//	    AddRaw(mictofun.Warning("Be careful with this operation."))
package mictofun

import (
	"strings"
)

// AlertType represents the type of alert/callout.
type AlertType int

const (
	AlertNote AlertType = iota
	AlertTip
	AlertImportant
	AlertWarning
	AlertCaution
)

// alertTypeStrings maps AlertType to its markdown string.
var alertTypeStrings = map[AlertType]string{
	AlertNote:      "NOTE",
	AlertTip:       "TIP",
	AlertImportant: "IMPORTANT",
	AlertWarning:   "WARNING",
	AlertCaution:   "CAUTION",
}

// Alert creates a GitHub-style alert callout.
//
// Output format:
//
//	> [!NOTE]
//	> Your content here
func Alert(alertType AlertType, content string) string {
	typeStr, ok := alertTypeStrings[alertType]
	if !ok {
		typeStr = "NOTE"
	}

	var sb strings.Builder
	sb.WriteString("> [!" + typeStr + "]\n")

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		sb.WriteString("> " + line + "\n")
	}

	return sb.String()
}

// Note creates a NOTE alert for general information.
func Note(content string) string {
	return Alert(AlertNote, content)
}

// Tip creates a TIP alert for helpful suggestions.
func Tip(content string) string {
	return Alert(AlertTip, content)
}

// Important creates an IMPORTANT alert for critical information.
func Important(content string) string {
	return Alert(AlertImportant, content)
}

// Warning creates a WARNING alert for potential issues.
func Warning(content string) string {
	return Alert(AlertWarning, content)
}

// Caution creates a CAUTION alert for dangerous actions.
func Caution(content string) string {
	return Alert(AlertCaution, content)
}

// Alert builder for method chaining on Doc

// Note adds a NOTE alert to the document.
func (d *Doc) Note(content string) *Doc {
	return d.write(Note(content))
}

// Tip adds a TIP alert to the document.
func (d *Doc) Tip(content string) *Doc {
	return d.write(Tip(content))
}

// Important adds an IMPORTANT alert to the document.
func (d *Doc) Important(content string) *Doc {
	return d.write(Important(content))
}

// Warning adds a WARNING alert to the document.
func (d *Doc) Warning(content string) *Doc {
	return d.write(Warning(content))
}

// Caution adds a CAUTION alert to the document.
func (d *Doc) Caution(content string) *Doc {
	return d.write(Caution(content))
}
