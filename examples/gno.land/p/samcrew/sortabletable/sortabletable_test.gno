package sortabletable

import (
	"net/url"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	// --- Case 1: Table with "Role" column sorted descending and param prefix "members-"
	table := &Table{
		Headings: []string{"Tier", "Member", "Role"},
		Rows: [][]string{
			{"T1", "g11111", "finance-officer"},
			{"T2", "g22222", "developer"},
			{"T3", "g33333", "developer"},
		},
	}

	u, _ := url.Parse("/test?members-sort-desc=Role")
	md := Render(u, table, "members-")

	expected := "| [Tier](/test?members-sort-desc=Tier) | [Member](/test?members-sort-desc=Member) | [Role ↓](/test?members-sort-asc=Role) |\n" +
		"| --- | --- | --- |\n" +
		"| T1 | g11111 | finance-officer |\n" +
		"| T3 | g33333 | developer |\n" +
		"| T2 | g22222 | developer |\n"

	// Trim spaces for comparison
	md = strings.TrimSpace(md)
	expected = strings.TrimSpace(expected)

	if md != expected {
		t.Errorf("Render() output mismatch.\nExpected:\n%s\nGot:\n%s", expected, md)
	}

	// --- Case 2: Just an invalid table (row with missing cell)
	table = &Table{
		Headings: []string{"Tier", "Member", "Role"},
		Rows: [][]string{
			{"T1", "g11111"}, // missing Role
		},
	}

	md = Render(u, table, "")
	expected = "sortabletable fails: row 1 has 2 cells, expected 3, because there are 3 columns.\n"

	if md != expected {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expected, md)
	}
}
