package outfmt

import (
	"testing"

	"gno.land/r/x/manfred_outfmt"
)

func TestRender(t *testing.T) {
	// home
	{
		got := outfmt.Render("")
		expected := `* [?fmt=stringer](/r/x/manfred_outfmt:?fmt=stringer)
* [?fmt=json](/r/x/manfred_outfmt:?fmt=json)
* [?fmt=jsonp](/r/x/manfred_outfmt:?fmt=jsonp)
`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}

	// stringer
	{
		got := outfmt.Render("?fmt=stringer")
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
		got := outfmt.Render("?fmt=stringer")
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
		got := outfmt.Render("?fmt=json")
		expected := `{"Number":746,"Text":"Hello Gnomes!","Numbers":[57,82,16,14,28,32]}`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}

	// jsonp
	{
		got := outfmt.Render("?fmt=jsonp")
		expected := `callback({"Number":795,"Text":"Hello Gnomes!","Numbers":[29,51,88,61,93,21,2,66,79]})`
		if got != expected {
			t.Errorf("expected %q, got %q.", expected, got)
		}
	}
}
