package gnopages

func init() {
	var (
		path  = "tokenomics"
		title = "Gno.land Tokenomics"
		// XXX: description = """
		body = `Lorem Ipsum`
	)
	_ = b.NewPost("", path, title, body, nil)
}
