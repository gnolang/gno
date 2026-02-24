package pages

func init() {
	var (
		path  = "tokenomics"
		title = "Gno.land Tokenomics"
		// XXX: description = """
		body = `Lorem Ipsum`
	)
	if err := b.NewPost("", path, title, body, "2022-05-20T13:17:28Z", nil, nil); err != nil {
		panic(err)
	}

}
