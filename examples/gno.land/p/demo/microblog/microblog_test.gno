package microblog

import (
	"log"
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
)

func TestMicroblog(t *testing.T) {
	const (
		title   string      = "test microblog"
		prefix  string      = "/r/test"
		author1 std.Address = testutils.TestAddress("author1")
		author2 std.Address = testutils.TestAddress("author2")
	)

	std.TestSetOrigCaller(author1)

	d := NewMicroblog(title, prefix)
	if d.Render("/wrongpath") != "404" {
		t.Fatalf("rendering not giving 404")
	}
	if d.Render("") == "404" {
		t.Fatalf("rendering / should not give 404")
	}
	if err := d.NewPage(); err != nil {
		t.Fatalf("%+v", err)
	}
	if _, err := d.GetPage(author1.String()); err != nil {
		t.Fatalf("silo should exist")
	}
	if _, err := d.GetPage("no such author"); err == nil {
		t.Fatalf("silo should not exist")
	}
	if err := d.NewPost("goodbyte, web2"); err != nil {
		t.Fatalf("could not create post")
	}

	std.TestSetOrigCaller(author2)

	if err := d.NewPost("hello, web3"); err != nil {
		t.Fatalf("could not create post")
	}
	if err := d.NewPost("hello again, web3"); err != nil {
		t.Fatalf("could not create post")
	}
	if err := d.NewPost("hi again, web4?"); err != nil {
		t.Fatalf("could not create post")
	}

	// println("--- MICROBLOG ---\n\n")
	// println(d.Render(""))
	// println("\n\n--- AUTHOR PAGE ---\n\n")
	// println(d.Render(author1.String()))
	// println("\n\n--- AUTHOR PAGE ---\n\n")
	// println(d.Render(author2.String()))
}
