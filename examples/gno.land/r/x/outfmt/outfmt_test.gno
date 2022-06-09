package outfmt

import (
	"testing"

	"gno.land/r/x/outfmt"
)

func TestRender(t *testing.T) {
	// home
	{
		got := outfmt.Render("")
		expected := `* [?fmt=stringer](/r/x/outfmt:?fmt=stringer)
* [?fmt=json](/r/x/outfmt:?fmt=json)
* [?fmt=jsonp](/r/x/outfmt:?fmt=jsonp)
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// stringer
	{
		got := outfmt.Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 602
Numbers: 38 88
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// stringer again (different seed)
	{
		got := outfmt.Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 418
Numbers: 41 0 14 67 91
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}

	// json
	{
		got := outfmt.Render("?fmt=json")
		expected := `{"Number":467,"Text":"Hello Gnomes!","Numbers":[8,62]}`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}

	// jsonp
	{
		got := outfmt.Render("?fmt=jsonp")
		expected := `callback({"Number":384,"Text":"Hello Gnomes!","Numbers":[15,78]})`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}
}
