package gnopages

func init() {
	path := "start"
	title := "Getting Started with Gno"
	// XXX: description := ""

	// TODO: codegen to use README files here

	/* TODO: port previous message: This is a demo of Gno smart contract programming.  This document was
	constructed by Gno onto a smart contract hosted on the data Realm
	name ["gno.land/r/demo/boards"](https://gno.land/r/demo/boards/)
	([github](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo/boards)).
	*/
	body := `## Getting Started with Gno

- [Install Gno Key](/r/demo/boards:testboard/5)
- TODO: add more links
`
	_ = b.NewPost("", path, title, body, "2022-05-20T13:17:28Z", nil, nil)
}
