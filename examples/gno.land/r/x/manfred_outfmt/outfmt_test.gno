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
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// stringer
	{
		got := outfmt.Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 957
Numbers: 3 54 32 88
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// stringer again (different seed)
	{
		got := outfmt.Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 141
Numbers: 98 27
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// json
	{
		got := outfmt.Render("?fmt=json")
		expected := `{"Number":801,"Text":"Hello Gnomes!","Numbers":[5,78,51,78,91,41]}`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// jsonp
	{
		got := outfmt.Render("?fmt=jsonp")
		expected := `callback({"Number":63,"Text":"Hello Gnomes!","Numbers":[2,66,50,73,81]})`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}
}
