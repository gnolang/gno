package outfmt

import (
	"testing"
)

func TestRender(t *testing.T) {
	// home
	{
		got := Render("")
		expected := `* [?fmt=stringer](/r/moul/outfmt:?fmt=stringer)
* [?fmt=json](/r/moul/outfmt:?fmt=json)
* [?fmt=jsonp](/r/moul/outfmt:?fmt=jsonp)
`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}

	// stringer
	{
		got := Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 222
Numbers: 34 44 39 7 72 48 74
`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}

	// stringer again (different seed)
	{
		got := Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 898
Numbers: 24 25 2
`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}

	// json
	{

		got := Render("?fmt=json")
		expected := `{"Number":746,"Text":"Hello Gnomes!","Numbers":[57,82,16,14,28,32]}`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}

	// jsonp
	{
		got := Render("?fmt=jsonp")
		expected := `callback({"Number":795,"Text":"Hello Gnomes!","Numbers":[29,51,88,61,93,21,2,66,79]})`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}
}
