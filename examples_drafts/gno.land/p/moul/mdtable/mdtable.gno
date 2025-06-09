// Package mdtable provides a simple way to create Markdown tables.
//
// Example usage:
//
//	import "gno.land/p/moul/mdtable"
//
//	func Render(path string) string {
//	    table := mdtable.Table{
//	        Headers: []string{"ID", "Title", "Status", "Date"},
//	    }
//	    table.Append([]string{"#1", "Add a new validator", "succeed", "2024-01-01"})
//	    table.Append([]string{"#2", "Change parameter", "timed out", "2024-01-02"})
//	    return table.String()
//	}
//
// Output:
//
//	| ID | Title | Status | Date |
//	| --- | --- | --- | --- |
//	| #1 | Add a new validator | succeed | 2024-01-01 |
//	| #2 | Change parameter | timed out | 2024-01-02 |
package mdtable

import (
	"strings"
)

type Table struct {
	Headers []string
	Rows    [][]string
	// XXX: optional headers alignment.
}

func (t *Table) Append(row []string) {
	t.Rows = append(t.Rows, row)
}

func (t Table) String() string {
	// XXX: switch to using text/tabwriter when porting to Gno to support
	// better-formatted raw Markdown output.

	if len(t.Headers) == 0 && len(t.Rows) == 0 {
		return ""
	}

	var sb strings.Builder

	if len(t.Headers) == 0 {
		t.Headers = make([]string, len(t.Rows[0]))
	}

	// Print header.
	sb.WriteString("| " + strings.Join(t.Headers, " | ") + " |\n")
	sb.WriteString("|" + strings.Repeat(" --- |", len(t.Headers)) + "\n")

	// Print rows.
	for _, row := range t.Rows {
		escapedRow := make([]string, len(row))
		for i, cell := range row {
			escapedRow[i] = strings.ReplaceAll(cell, "|", "&#124;") // Escape pipe characters.
		}
		sb.WriteString("| " + strings.Join(escapedRow, " | ") + " |\n")
	}

	return sb.String()
}
