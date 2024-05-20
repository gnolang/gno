package microblog

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/microblog"
	"gno.land/p/demo/testutils"
)

func TestMicroblog(t *testing.T) {
	const (
		author1 std.Address = testutils.TestAddress("author1")
		author2 std.Address = testutils.TestAddress("author2")
	)

	std.TestSetOrigCaller(author1)

	if Render("/wrongpath") != "404" {
		t.Fatalf("rendering not giving 404")
	}
	if Render("") == "404" {
		t.Fatalf("rendering / should not give 404")
	}
	if err := m.NewPost("goodbyte, web2"); err != nil {
		t.Fatalf("could not create post")
	}
	if _, err := m.GetPage(author1.String()); err != nil {
		t.Fatalf("silo should exist")
	}
	if _, err := m.GetPage("no such author"); err == nil {
		t.Fatalf("silo should not exist")
	}

	std.TestSetOrigCaller(author2)

	if err := m.NewPost("hello, web3"); err != nil {
		t.Fatalf("could not create post")
	}
	if err := m.NewPost("hello again, web3"); err != nil {
		t.Fatalf("could not create post")
	}
	if err := m.NewPost("hi again,\n web4?"); err != nil {
		t.Fatalf("could not create post")
	}

	println("--- MICROBLOG ---\n\n")
	if rendering := Render(""); rendering != `# gno-based microblog

# pages

- [g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6](/r/demo/microblog:g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6)
- [g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00](/r/demo/microblog:g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00)
` {
		t.Fatalf("incorrect rendering /: '%s'", rendering)
	}

	if rendering := strings.TrimSpace(Render(author1.String())); rendering != `## [g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6](/r/demo/microblog:g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6)

joined 2009-13-02, last updated 2009-13-02

## feed

> goodbyte, web2
>
> *Fri, 13 Feb 2009 23:31:30 UTC*` {
		t.Fatalf("incorrect rendering /: '%s'", rendering)
	}

	if rendering := strings.TrimSpace(Render(author2.String())); rendering != `## [g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00](/r/demo/microblog:g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00)

joined 2009-13-02, last updated 2009-13-02

## feed

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
