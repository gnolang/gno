// Section provides layout components like collapsible sections and columns.
//
// These use HTML tags that are supported by gno.land's markdown renderer.
//
// Example usage:
//
//	doc := mictofun.New().
//	    H1("FAQ").
//	    AddRaw(mictofun.Collapsible("What is Gno?", "Gno is a smart contract platform...")).
//	    AddRaw(mictofun.Collapsible("How to get started?", "Check the docs at..."))
package mictofun

import (
	"strings"
)

// Collapsible creates a collapsible/expandable section using HTML details tag.
//
// Output format:
//
//	<details><summary>Title</summary>
//
//	Content here
//	</details>
func Collapsible(title, content string) string {
	return "<details><summary>" + EscapeText(title) + "</summary>\n\n" + content + "\n</details>\n"
}

// CollapsibleOpen creates a collapsible section that is open by default.
func CollapsibleOpen(title, content string) string {
	return "<details open><summary>" + EscapeText(title) + "</summary>\n\n" + content + "\n</details>\n"
}

// Collapsible adds a collapsible section to the document.
func (d *Doc) Collapsible(title, content string) *Doc {
	return d.write(Collapsible(title, content))
}

// CollapsibleOpen adds an open collapsible section to the document.
func (d *Doc) CollapsibleOpen(title, content string) *Doc {
	return d.write(CollapsibleOpen(title, content))
}

// ============================================================================
// COLUMNS (gno.land specific)
// ============================================================================

// Columns creates a multi-column layout using gno.land's <gno-columns> tag.
// Each string in cols represents content for one column.
// Maximum 4 columns are supported per row.
//
// Output format:
//
//	<gno-columns>
//	Column 1 content
//	|||
//	Column 2 content
//	</gno-columns>
func Columns(cols ...string) string {
	if len(cols) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<gno-columns>\n")

	for i, col := range cols {
		if i > 0 {
			sb.WriteString("|||\n")
		}
		sb.WriteString(col + "\n")
	}

	sb.WriteString("</gno-columns>\n")
	return sb.String()
}

// ColumnsN splits content into rows of N columns each.
// Useful for creating grid layouts with consistent column counts.
func ColumnsN(content []string, colsPerRow int) string {
	if len(content) == 0 || colsPerRow <= 0 {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < len(content); i += colsPerRow {
		end := i + colsPerRow
		if end > len(content) {
			end = len(content)
		}
		row := content[i:end]
		sb.WriteString(Columns(row...))
	}
	return sb.String()
}

// Cols adds a multi-column layout to the document.
func (d *Doc) Cols(cols ...string) *Doc {
	return d.write(Columns(cols...))
}

// ============================================================================
// BADGES & SHIELDS
// ============================================================================

// Badge creates a shields.io style badge.
func Badge(label, message, color string) string {
	url := "https://img.shields.io/badge/" + label + "-" + message + "-" + color
	return Image(label+": "+message, url)
}

// BadgeLink creates a clickable badge.
func BadgeLink(label, message, color, linkURL string) string {
	badgeURL := "https://img.shields.io/badge/" + label + "-" + message + "-" + color
	return ImageWithLink(label+": "+message, badgeURL, linkURL)
}

// ============================================================================
// DIVIDERS & SPACING
// ============================================================================

// Spacer adds vertical spacing using line breaks.
func Spacer(lines int) string {
	if lines <= 0 {
		lines = 1
	}
	return strings.Repeat("<br>\n", lines)
}

// Spacer adds vertical spacing to the document.
func (d *Doc) Spacer(lines int) *Doc {
	return d.write(Spacer(lines))
}

// Center creates centered text using HTML.
func Center(content string) string {
	return "<div align=\"center\">\n\n" + content + "\n\n</div>\n"
}

// Center adds centered content to the document.
func (d *Doc) Center(content string) *Doc {
	return d.write(Center(content))
}

// Right creates right-aligned text using HTML.
func Right(content string) string {
	return "<div align=\"right\">\n\n" + content + "\n\n</div>\n"
}

// Right adds right-aligned content to the document.
func (d *Doc) Right(content string) *Doc {
	return d.write(Right(content))
}
