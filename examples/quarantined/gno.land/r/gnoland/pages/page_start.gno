package pages

func init() {
	path := "start"
	title := "Getting Started with Gno"
	body := `## Getting Started with Gno

Visit the [getting-started](https://github.com/gnolang/getting-started) repo to try Gno in 5 minutes!
`
	if err := b.NewPost("", path, title, body, "2022-05-20T13:17:26Z", nil, nil); err != nil {
		panic(err)
	}
}
