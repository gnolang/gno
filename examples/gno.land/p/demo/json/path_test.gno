package json

import "testing"

func TestParseJSONPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{name: "Empty string path", path: "", expected: []string{}},
		{name: "Root only path", path: "$", expected: []string{"$"}},
		{name: "Root with dot path", path: "$.", expected: []string{"$"}},
		{name: "All objects in path", path: "$..", expected: []string{"$", ".."}},
		{name: "Only children in path", path: "$.*", expected: []string{"$", "*"}},
		{name: "All objects' children in path", path: "$..*", expected: []string{"$", "..", "*"}},
		{name: "Simple dot notation path", path: "$.root.element", expected: []string{"$", "root", "element"}},
		{name: "Complex dot notation path with wildcard", path: "$.root.*.element", expected: []string{"$", "root", "*", "element"}},
		{name: "Path with array wildcard", path: "$.phoneNumbers[*].type", expected: []string{"$", "phoneNumbers", "*", "type"}},
		{name: "Path with filter expression", path: "$.store.book[?(@.price < 10)].title", expected: []string{"$", "store", "book", "?(@.price < 10)", "title"}},
		{name: "Path with formula", path: "$..phoneNumbers..('ty' + 'pe')", expected: []string{"$", "..", "phoneNumbers", "..", "('ty' + 'pe')"}},
		{name: "Simple bracket notation path", path: "$['root']['element']", expected: []string{"$", "'root'", "'element'"}},
		{name: "Complex bracket notation path with wildcard", path: "$['root'][*]['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "Bracket notation path with integer index", path: "$['store']['book'][0]['title']", expected: []string{"$", "'store'", "'book'", "0", "'title'"}},
		{name: "Complex path with wildcard in bracket notation", path: "$['root'].*['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "Mixed notation path with dot after bracket", path: "$.['root'].*.['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "Mixed notation path with dot before bracket", path: "$['root'].*.['element']", expected: []string{"$", "'root'", "*", "'element'"}},
		{name: "Single character path with root", path: "$.a", expected: []string{"$", "a"}},
		{name: "Multiple characters path with root", path: "$.abc", expected: []string{"$", "abc"}},
		{name: "Multiple segments path with root", path: "$.a.b.c", expected: []string{"$", "a", "b", "c"}},
		{name: "Multiple segments path with wildcard and root", path: "$.a.*.c", expected: []string{"$", "a", "*", "c"}},
		{name: "Multiple segments path with filter and root", path: "$.a[?(@.b == 'c')].d", expected: []string{"$", "a", "?(@.b == 'c')", "d"}},
		{name: "Complex path with multiple filters", path: "$.a[?(@.b == 'c')].d[?(@.e == 'f')].g", expected: []string{"$", "a", "?(@.b == 'c')", "d", "?(@.e == 'f')", "g"}},
		{name: "Complex path with multiple filters and wildcards", path: "$.a[?(@.b == 'c')].*.d[?(@.e == 'f')].g", expected: []string{"$", "a", "?(@.b == 'c')", "*", "d", "?(@.e == 'f')", "g"}},
		{name: "Path with array index and root", path: "$.a[0].b", expected: []string{"$", "a", "0", "b"}},
		{name: "Path with multiple array indices and root", path: "$.a[0].b[1].c", expected: []string{"$", "a", "0", "b", "1", "c"}},
		{name: "Path with array index, wildcard and root", path: "$.a[0].*.c", expected: []string{"$", "a", "0", "*", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reult, _ := ParsePath(tt.path)
			if !isEqualSlice(reult, tt.expected) {
				t.Errorf("ParsePath(%s) expected: %v, got: %v", tt.path, tt.expected, reult)
			}
		})
	}
}

func isEqualSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
