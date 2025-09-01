// Package sortabletable provides functionality to render a Markdown table with sortable columns.
// It allows users to click on column headers to sort the table in ascending or descending sort direction.
// The sorting state is managed via URL query parameters.
// It displays an error if the table is malformed (e.g. rows with missing cells).
// Multiple sortabletable can be rendered on the same page by using a paramPrefix for each Render.
package sortabletable

import (
	"net/url"
	"sort"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

// Table holds the headings and rows for rendering.
// Each row must have the same number of cells as there are headings.
type Table struct {
	Headings []string   // ["A", "B", "C"]
	Rows     [][]string // [["a1","b1","c1"], ["a2","b2","c2"], ...]
}

// Render generates a Markdown table from a Table struct with sortable columns based on URL params.
// paramPrefix is an optional prefix for in the URL to identify the sortabletable Renders (e.g. "members-").
func Render(u *url.URL, table *Table, paramPrefix string) string {
	direction := ""
	currentHeading := ""
	if h := u.Query().Get(paramPrefix + "sort-asc"); h != "" {
		direction = "asc"
		currentHeading = h
	} else if h := u.Query().Get(paramPrefix + "sort-desc"); h != "" {
		direction = "desc"
		currentHeading = h
	}

	var sb strings.Builder

	// Find the index of the column to sort
	colIndex := -1
	for i, h := range table.Headings {
		if h == currentHeading {
			colIndex = i
			break
		}
	}

	// Sort rows if necessary
	if colIndex != -1 {
		sortRows(table.Rows, colIndex, direction == "asc")
	}

	// Build header
	sb.WriteString(buildHeader(u, table.Headings, currentHeading, direction, paramPrefix))
	sb.WriteString("\n")

	numCols := len(table.Headings)

	// Build rows
	for i, row := range table.Rows {
		// Validate row length
		if len(row) != numCols {
			return "sortabletable fails: row " + ufmt.Sprintf("%d", i+1) + " has " +
				ufmt.Sprintf("%d", len(row)) + " cells, expected " +
				ufmt.Sprintf("%d", numCols) + ", because there are " + ufmt.Sprintf("%d", numCols) + " columns.\n"
		}

		sb.WriteString("|")
		for _, cell := range row {
			sb.WriteString(" " + cell + " |")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildHeader builds the Markdown header row with clickable links and arrows
func buildHeader(u *url.URL, headings []string, currentHeading, direction string, paramPrefix string) string {
	var sb strings.Builder
	sb.WriteString("|")
	for _, h := range headings {
		arrow := ""
		if h == currentHeading {
			if direction == "asc" {
				arrow = " ↑"
			} else if direction == "desc" {
				arrow = " ↓"
			}
		}

		// Build URL for the header link with toggle logic
		newURL := *u
		q := newURL.Query()
		if h == currentHeading {
			// Toggle sort direction
			if direction == "asc" {
				q.Del(paramPrefix + "sort-asc")
				q.Set(paramPrefix+"sort-desc", h)
			} else {
				q.Del(paramPrefix + "sort-desc")
				q.Set(paramPrefix+"sort-asc", h)
			}
		} else {
			// First click defaults to descending
			q.Del(paramPrefix + "sort-asc")
			q.Set(paramPrefix+"sort-desc", h)
		}
		newURL.RawQuery = q.Encode()
		link := md.Link(h+arrow, newURL.String())
		sb.WriteString(" " + link + " |")
	}

	sb.WriteString("\n|")
	for range headings {
		sb.WriteString(" --- |")
	}
	return sb.String()
}

// rowSorter implements sort.Interface for sorting rows by a specific column.
type rowSorter struct {
	rows      [][]string
	colIndex  int
	ascending bool
}

func (rs rowSorter) Len() int {
	return len(rs.rows)
}

func (rs rowSorter) Less(i, j int) bool {
	iCell := rs.rows[i][rs.colIndex]
	jCell := rs.rows[j][rs.colIndex]
	if rs.ascending {
		return iCell < jCell
	}
	return iCell > jCell
}

func (rs rowSorter) Swap(i, j int) {
	rs.rows[i], rs.rows[j] = rs.rows[j], rs.rows[i]
}

// sortRows sorts the rows slice by a given column index and direction
func sortRows(rows [][]string, colIndex int, ascending bool) {
	sort.Sort(rowSorter{rows, colIndex, ascending})
}
