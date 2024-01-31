package gnopages

func init() {
	path := "partners"
	title := "Partners"
	// XXX: description := """
	body := `## Partnerships

### Fund and Grants Program

Are you a builder, tinkerer, or researcher? If you’re looking to create awesome dApps, tooling, infrastructure, or smart contract libraries on Gno.land, you can apply for a grant. The Gno.land Ecosystem Fund and Grants program provides financial contributions for individuals and teams to innovate on the platform.

<div class="stack">

[More information here](https://github.com/gnolang/ecosystem-fund-grants)

</div><!-- end stack-->
`

	_ = b.NewPost("", path, title, body, nil)
}
