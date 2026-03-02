package commondao

import (
	"strings"
	"testing"
)

func TestRender_StaticHome(t *testing.T) {
	output := Render("")
	expected := []string{
		"# A CommonDao Example",
		"This realm is meant to demonstrate how to use the [/p/nt/commondao](/p/nt/commondao) pure package.",
		"The commondao package is meant to act as a DAO factory to use in any realms easily.",
		"In the source code, you'll find out how to create a basic DAO and where to look for more informations about more advanced options",
		"## DAO Overview",
		"**Title:** CommonDAO Documentation Realm",
		"**Description:** An example DAO",
		"### DAO's members",
		"### DAO's Active Proposals",
	}

	for _, snippet := range expected {
		if !strings.Contains(output, snippet) {
			t.Errorf("Output missing expected snippet:\n%s", snippet)
		}
	}
}

func TestRender_JoinURL(t *testing.T) {
	output := Render("join")
	expected := []string{
		"# A CommonDao Example",
		"## You found the secret of this realm!",
		"Well done on reading the documentation of the CommonDAO package!",
		"You can request to joins the DAO by clicking the button below:",
	}

	for _, snippet := range expected {
		if !strings.Contains(output, snippet) {
			t.Errorf("Output for 'join' URL missing expected snippet:\n%s", snippet)
		}
	}
}
