package gnopages

func init() {
	path := "start"
	title := "Getting Started with Gno"
	body := `## Getting Started with Gno

Visit the [getting-started](https://github.com/gnolang/getting-started) repo to try Gno in 5 minutes!
`
	_ = b.NewPost("", path, title, body, "2022-05-20T13:17:28Z", nil, nil)
}
