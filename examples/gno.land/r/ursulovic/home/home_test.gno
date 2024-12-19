package home

import (
	"std"
	"testing"
)

func TestUpdatePfp(t *testing.T) {
	owner := std.Address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	std.TestSetOrigCaller(owner)

	profile.pfp = "abc"

	UpdatePfp("https://i.ibb.co/W28NPkw/beograd.webp")

	if pfp != "https://i.ibb.co/W28NPkw/beograd.webp" {
		t.Fatalf("Excpected result: https://i.ibb.co/W28NPkw/beograd.webp, got %s", pfp)
	}
}

func TestUpdateAboutMe(t *testing.T) {
	owner := std.Address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	std.TestSetOrigCaller(owner)

	aboutMe = ""

	excpected := "Testing about me..."

	UpdateAboutMe(excpected)

	if aboutMe != excpected {
		t.Fatalf("Excpected about me to be %s, got %s", excpected, aboutMe)
	}
}

func TestUpdatePhrase(t *testing.T) {
	owner := std.Address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	std.TestSetOrigCaller(owner)

	index := 0
	newPhrase := "New test phrase"

	UpdatePhrase(index, newPhrase)

	if phrases[index] != newPhrase {
		t.Fatalf("Expected phrase at index %d to be %s, got %s", index, newPhrase, phrases[index])
	}
}

func TestRender(t *testing.T) {
	owner := std.Address("g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x")
	std.TestSetOrigCaller(owner)

	std.TestSetHeight(0)
	rendered := Render("")
	if !strings.Contains(rendered, phrases[0]) {
		t.Fatalf("Expected phrase at index 0 to be displayed, but got %s", rendered)
	}
}
