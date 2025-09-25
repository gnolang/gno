// Package tablesort provides functionality to render a Markdown table with sortable columns.
// It allows users to click on column headers to sort the table in ascending or descending sort direction.
// The sorting state is managed via URL query parameters.
// It displays an error if the table is malformed (e.g. rows with missing cells).
// Multiple tablesort can be rendered on the same page by using a paramPrefix for each Render (See the Render function).
package tablesort

import (
	"sort"
)

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

// SortRows sorts the rows slice by a given column index and direction
func SortRows(rows [][]string, colIndex int, ascending bool) {
	sort.Sort(rowSorter{rows, colIndex, ascending})
}
