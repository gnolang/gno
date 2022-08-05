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
Number: 250
Numbers: 98 25 21 45 33 32 81
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// stringer again (different seed)
	{
		got := outfmt.Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 550
Numbers: 19 31 35 94 76 63 66 20 48
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}

	// json
	{
		got := outfmt.Render("?fmt=json")
		expected := `{"Number":100,"Text":"Hello Gnomes!","Numbers":[66,24,80]}`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}

	// jsonp
	{
		got := outfmt.Render("?fmt=jsonp")
		expected := `callback({"Number":717,"Text":"Hello Gnomes!","Numbers":[60,1,43,24,1,21]})`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}
}
