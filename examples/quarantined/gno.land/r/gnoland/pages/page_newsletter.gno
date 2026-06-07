package gnopages

import (
	"gno.land/p/leon/svgbtn"
)

func init() {
	path := "newsletter"
	title := "Sign up for Gno News"
	body := "**Stay in the Gno by signing up for our newsletter. You'll get the scoop on dev updates, fresh content, and community news.**\n\n"
	body += svgbtn.Button(
		256,
		44,
		"#226c57",
		"#ffffff",
		"Subscribe to stay in the Gno",
		"https://land.us18.list-manage.com/subscribe?u=8befe3303cf82796d2c1a1aff&id=271812000b",
	)

	if err := b.NewPost("", path, title, body, "2025-07-14T00:00:00Z", nil, nil); err != nil {
		panic(err)
	}
}
