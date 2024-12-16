package avl_pager

import (
	"testing"
)

func TestRender(t *testing.T) {
	// Test default Render output (first page)
	output := Render("")
	expected := `# Paginated Items
Page 1 of 10

**1** | [2](?page=2) | [3](?page=3) | … | [10](?page=10)

- Item1: Value of Item1
- Item10: Value of Item10
- Item100: Value of Item100
- Item11: Value of Item11
- Item12: Value of Item12
- Item13: Value of Item13
- Item14: Value of Item14
- Item15: Value of Item15
- Item16: Value of Item16
- Item17: Value of Item17

**1** | [2](?page=2) | [3](?page=3) | … | [10](?page=10)`
	if output != expected {
		t.Errorf("Render(\"\") failed, got:\n%s", output)
	}
}

func TestRender_page2(t *testing.T) {
	// Test Render output for a custom page (page 2)
	output := Render("?page=2&size=10")
	expected := `# Paginated Items
Page 2 of 10

[1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | … | [10](?page=10)

- Item18: Value of Item18
- Item19: Value of Item19
- Item2: Value of Item2
- Item20: Value of Item20
- Item21: Value of Item21
- Item22: Value of Item22
- Item23: Value of Item23
- Item24: Value of Item24
- Item25: Value of Item25
- Item26: Value of Item26

[1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | … | [10](?page=10)`
	if output != expected {
		t.Errorf("Render(\"\") failed, got:\n%s", output)
	}
}
