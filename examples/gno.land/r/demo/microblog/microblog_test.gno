package microblog

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestMicroblog(t *testing.T) {
	var (
		author1 address = testutils.TestAddress("author1")
		author2 address = testutils.TestAddress("author2")
	)

	testing.SetOriginCaller(author1)

	urequire.Equal(t, "404", Render("/wrongpath"), "rendering not giving 404")
	urequire.NotEqual(t, "404", Render(""), "rendering / should not give 404")
	urequire.NoError(t, m.NewPost("goodbyte, web2"), "could not create post")

	_, err := m.GetPage(author1.String())
	urequire.NoError(t, err, "silo should exist")

	_, err = m.GetPage("no such author")
	urequire.Error(t, err, "silo should not exist")

	testing.SetOriginCaller(author2)

	urequire.NoError(t, m.NewPost("hello, web3"), "could not create post")
	urequire.NoError(t, m.NewPost("hello again, web3"), "could not create post")
	urequire.NoError(t, m.NewPost("hi again,\n web4?"), "could not create post")

	println("--- MICROBLOG ---\n\n")

	expected := `# gno-based microblog

# pages

- [g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6](/r/demo/microblog:g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6)
- [g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00](/r/demo/microblog:g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00)
`
	urequire.Equal(t, expected, Render(""), "incorrect rendering")

	expected = `## [g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6](/r/demo/microblog:g1v96hg6r0wgc47h6lta047h6lta047h6lm33tq6)

joined 2009-13-02, last updated 2009-13-02

## feed

> goodbyte, web2
>
> *Fri, 13 Feb 2009 23:31:30 UTC*`

	urequire.Equal(t, expected, strings.TrimSpace(Render(author1.String())), "incorrect rendering")

	expected = `## [g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00](/r/demo/microblog:g1v96hg6r0wge97h6lta047h6lta047h6lyz7c00)

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
> *Fri, 13 Feb 2009 23:31:30 UTC*`

	urequire.Equal(t, expected, strings.TrimSpace(Render(author2.String())), "incorrect rendering")
}
