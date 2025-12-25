// Table provides a fluent API for creating markdown tables with alignment support.
//
// Example usage:
//
//	table := mictofun.NewTable().
//	    Headers("Name", "Amount", "Status").
//	    SetAlign(mictofun.AlignLeft, mictofun.AlignRight, mictofun.AlignCenter).
//	    Row("Alice", "$100", "✓").
//	    Row("Bob", "$250", "Pending")
//
//	doc.Add(table)
package mictofun

import (
	"strings"
)

// Align represents column alignment in tables.
type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// Table is a fluent builder for creating markdown tables.
type Table struct {
	headers []string
	aligns  []Align
	rows    [][]string
}

// NewTable creates a new table builder.
func NewTable() *Table {
	return &Table{
		headers: make([]string, 0),
		aligns:  make([]Align, 0),
		rows:    make([][]string, 0),
	}
}

// Headers sets the table headers.
func (t *Table) Headers(cols ...string) *Table {
	t.headers = cols
	return t
}

// SetAlign sets column alignments. Use AlignLeft, AlignCenter, or AlignRight constants.
func (t *Table) SetAlign(aligns ...Align) *Table {
	t.aligns = aligns
	return t
}

// Row adds a single row to the table.
func (t *Table) Row(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Rows adds multiple rows at once.
func (t *Table) Rows(rows [][]string) *Table {
	t.rows = append(t.rows, rows...)
	return t
}

// String returns the markdown representation of the table.
func (t *Table) String() string {
	if len(t.headers) == 0 && len(t.rows) == 0 {
		return ""
	}

	var sb strings.Builder

	// Determine column count
	colCount := len(t.headers)
	if colCount == 0 && len(t.rows) > 0 {
		colCount = len(t.rows[0])
		t.headers = make([]string, colCount) // Empty headers
	}

	// Write header row
	sb.WriteString("| ")
	for i, h := range t.headers {
		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString(escapeTableCell(h))
	}
	sb.WriteString(" |\n")

	// Write separator row with alignment
	sb.WriteString("|")
	for i := 0; i < colCount; i++ {
		align := AlignLeft
		if i < len(t.aligns) {
			align = t.aligns[i]
		}
		switch align {
		case AlignCenter:
			sb.WriteString(" :---: |")
		case AlignRight:
			sb.WriteString(" ---: |")
		default:
			sb.WriteString(" --- |")
		}
	}
	sb.WriteString("\n")

	// Write data rows
	for _, row := range t.rows {
		sb.WriteString("| ")
		for i := 0; i < colCount; i++ {
			if i > 0 {
				sb.WriteString(" | ")
			}
			if i < len(row) {
				sb.WriteString(escapeTableCell(row[i]))
			}
		}
		sb.WriteString(" |\n")
	}

	return sb.String()
}

// escapeTableCell escapes pipe characters in table cells.
func escapeTableCell(cell string) string {
	return strings.ReplaceAll(cell, "|", "&#124;")
}

// TableFromData creates a table from a 2D string slice.
// First row is used as headers.
func TableFromData(data [][]string) *Table {
	t := NewTable()
	if len(data) == 0 {
		return t
	}
	t.Headers(data[0]...)
	if len(data) > 1 {
		t.Rows(data[1:])
	}
	return t
}

// SimpleTable creates a quick table with headers and rows.
func SimpleTable(headers []string, rows [][]string) string {
	return NewTable().Headers(headers...).Rows(rows).String()
}
