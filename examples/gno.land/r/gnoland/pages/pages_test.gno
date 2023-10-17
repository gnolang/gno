package gnopages

import (
	"std"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	printedOnce := false
	got := Render("")
	expectedSubtrings := []string{
		"/r/gnoland/pages:p/events",
		"/r/gnoland/pages:p/tokenomics",
		"/r/gnoland/pages:p/start",
		"/r/gnoland/pages:p/gor",
		"/r/gnoland/pages:p/about",
		"/r/gnoland/pages:p/language",
	}
	for _, substring := range expectedSubtrings {
		if !strings.Contains(got, substring) {
			if !printedOnce {
				println(got)
				printedOnce = true
			}
			t.Errorf("expected %q, but not found.", substring)
		}
	}
}

func TestAbout(t *testing.T) {
	printedOnce := false
	got := Render("p/about")
	expectedSubtrings := []string{
		"# About Gno.land",
		"Gno.land is a platform to write smart contracts in Gnolang (Gno).",
	}
	for _, substring := range expectedSubtrings {
		if !strings.Contains(got, substring) {
			if !printedOnce {
				println(got)
				printedOnce = true
			}
			t.Errorf("expected %q, but not found.", substring)
		}
	}
}
