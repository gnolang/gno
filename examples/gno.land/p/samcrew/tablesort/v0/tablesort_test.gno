package tablesort

import (
	"net/url"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	// --- Case 1: Testing Render with a Table with "Role" column sorted descending and param prefix "members-"
	table := &Table{
		Headings: []string{"Tier", "Member", "Role"},
		Rows: [][]string{
			{"T1", "g11111", "finance-officer"},
			{"T2", "g22222", "developer"},
			{"T3", "g33333", "developer"},
		},
	}

	u, _ := url.Parse("/test?members-sort-desc=Tier")
	md := Render(u, table, "members-")

	expected := "| [Tier ↓](/test?members-sort-asc=Tier) | [Member](/test?members-sort-desc=Member) | [Role](/test?members-sort-desc=Role) |\n" +
		"| --- | --- | --- |\n" +
		"| T3 | g33333 | developer |\n" +
		"| T2 | g22222 | developer |\n" +
		"| T1 | g11111 | finance-officer |\n"

	// Trim spaces for comparison
	md = strings.TrimSpace(md)
	expected = strings.TrimSpace(expected)

	if md != expected {
		t.Errorf("Render() output mismatch.\nExpected:\n%s\nGot:\n%s", expected, md)
	}

	// --- Case 2: Testing Render with an invalid Table (row with missing cell)
	table = &Table{
		Headings: []string{"Tier", "Member", "Role"},
		Rows: [][]string{
			{"T1", "g11111"}, // Missing the "Role" cell
		},
	}

	md = Render(u, table, "")
	expected = "tablesort fails: row 1 has 2 cells, expected 3, because there are 3 columns.\n"

	if md != expected {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expected, md)
	}

	// --- Case 3: Testing SortRows
	rows := [][]string{
		{"T1", "g11111", "finance-officer"},
		{"T2", "g22222", "developer"},
		{"T3", "g33333", "developer"},
	}

	SortRows(rows, 2, false) // Sort by "Role" descending
	expected = "finance-officer"

	if rows[0][2] != expected {
		t.Errorf("SortRows() failed. Expected first role: %s, got: %s", expected, rows[0][2])
	}
}
