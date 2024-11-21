package hello

import (
	"testing"
)

func TestHello(t *testing.T) {
	expected := "# Hello, 世界！"
	got := Render("")
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}

	got = Render("world")
	expected = "# Hello, world!"
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}
