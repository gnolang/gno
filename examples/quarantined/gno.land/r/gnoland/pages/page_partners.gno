package pages

func init() {
	path := "partners"
	title := "Partnerships"
	body := `### Fund and Grants Program

Are you a builder, tinkerer, or researcher? If you’re looking to create awesome dApps, tooling, infrastructure, 
or smart contract libraries on Gno.land, you can apply for a grant. The Gno.land Ecosystem Fund and Grants program 
provides financial contributions for individuals and teams to innovate on the platform.

Read more about our Funds and Grants program [here](https://github.com/gnolang/grants).
`
	if err := b.NewPost("", path, title, body, "2022-05-20T13:17:25Z", nil, nil); err != nil {
		panic(err)
	}
}
