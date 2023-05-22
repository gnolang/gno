package microblog

import (
	"log"
	"std"
	"strings"
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
	if err := d.NewPost("goodbyte, web2"); err != nil {
		t.Fatalf("could not create post")
	}
	if _, err := d.GetPage(author1.String()); err != nil {
		t.Fatalf("silo should exist")
	}
	if _, err := d.GetPage("no such author"); err == nil {
		t.Fatalf("silo should not exist")
	}

	std.TestSetOrigCaller(author2)

	if err := d.NewPost("hello, web3"); err != nil {
		t.Fatalf("could not create post")
	}
	if err := d.NewPost("hello again, web3"); err != nil {
		t.Fatalf("could not create post")
	}
	if err := d.NewPost("hi again,\n web4?"); err != nil {
		t.Fatalf("could not create post")
	}

	println("--- MICROBLOG ---\n\n")
	if rendering := d.Render(""); rendering != `# test microblog

# pages

- [g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6](/r/testg1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6)
- [g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00](/r/testg1v96hg6r0wge97h6lta047h6lta047h6lyz7c00)
` {
		t.Fatalf("incorrect rendering /: '%s'", rendering)
	}

	if rendering := strings.TrimSpace(d.Render(author1.String())); rendering != `# g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6

joined 2009-13-02, last updated 2009-13-02

## posts

> goodbyte, web2
>
> *Fri, 13 Feb 2009 23:31:30 UTC*` {
		t.Fatalf("incorrect rendering /: '%s'", rendering)
	}

	if rendering := strings.TrimSpace(d.Render(author2.String())); rendering != `# g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00

joined 2009-13-02, last updated 2009-13-02

## posts

> hi again,
>
> web4?
>
> *Fri, 13 Feb 2009 23:31:30 UTC*

> hello again, web3
>
> *Fri, 13 Feb 2009 23:31:30 UTC*

> hello, web3
>
> *Fri, 13 Feb 2009 23:31:30 UTC*` {
		t.Fatalf("incorrect rendering /: '%s'", rendering)
	}
}
