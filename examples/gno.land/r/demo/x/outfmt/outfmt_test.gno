package outfmt

import (
	"testing"

	"gno.land/r/demo/x/outfmt"
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
Number: 799
Numbers: 18 54 54 9 23 55 31
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}

	// stringer again (different seed)
	{
		got := outfmt.Render("?fmt=stringer")
		expected := `Text: Hello Gnomes!
Number: 618
Numbers: 25 64 25 19
`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}

	// json
	{
		got := outfmt.Render("?fmt=json")
		expected := `{"Number":39,"Text":"Hello Gnomes!","Numbers":[89,89,70,53,68,88,27,40,79]}`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}

	// jsonp
	{
		got := outfmt.Render("?fmt=jsonp")
		expected := `callback({"Number":253,"Text":"Hello Gnomes!","Numbers":[11,12]})`
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}

	}
}
