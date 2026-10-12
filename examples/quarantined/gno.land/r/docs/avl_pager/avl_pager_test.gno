package avl_pager

import (
	"testing"
)

func TestRender(t *testing.T) {
	req := "?sort=name&order=asc"
	output := Render(req)
	expected := `# Paginated Items
Page 1 of 10

**1** | [2](?page=2&order=asc&sort=name) | [3](?page=3&order=asc&sort=name) | … | [10](?page=10&order=asc&sort=name)

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

**1** | [2](?page=2&order=asc&sort=name) | [3](?page=3&order=asc&sort=name) | … | [10](?page=10&order=asc&sort=name)`
	if output != expected {
		t.Errorf("Render(%q) failed, got:\n%s\nwant:\n%s", req, output, expected)
	}
}

func TestRender_page2(t *testing.T) {
	req := "?page=2&size=10&sort=name&order=asc"
	output := Render(req)
	expected := `# Paginated Items
Page 2 of 10

[1](?page=1&order=asc&size=10&sort=name) | **2** | [3](?page=3&order=asc&size=10&sort=name) | [4](?page=4&order=asc&size=10&sort=name) | … | [10](?page=10&order=asc&size=10&sort=name)

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

[1](?page=1&order=asc&size=10&sort=name) | **2** | [3](?page=3&order=asc&size=10&sort=name) | [4](?page=4&order=asc&size=10&sort=name) | … | [10](?page=10&order=asc&size=10&sort=name)`
	if output != expected {
		t.Errorf("Render(%q) failed, got:\n%s\nwant:\n%s", req, output, expected)
	}
}
