package gnopages

func init() {
	path := "links"
	title := "Gno.land Linker"
	body := `

Below are Gno.land's most important links.

#### [Gno GitHub, the project's main repository](https://github.com/gnolang/gno)
#### [Getting Started with Gno](https://github.com/gnolang/getting-started/)
#### [Gno.land Official Docs](https://docs.gno.land/)
#### [Gnoverse Github, a community organization for Gnomes](https://github.com/gnoverse)
#### [Gno.land Workshops, a repo for Gno resources](https://github.com/gnolang/workshops)
#### [Gno.land Students Program](https://github.com/gnolang/student-contributors-program/)
#### [Gno.land Contributor page](https://gno.land/contribute)
#### [Gno.land Discord](https://discord.gg/gnoland)
#### [Gno.land X](https://x.com/_gnoland)
#### [Gno.land YouTube](https://www.youtube.com/@_gnoland)
#### [Gno.land Playground, an online editor for exploring Gno.land](https://play.gno.land)
#### [GnoScan, a Gno.land block explorer](https://gnoscan.io/)
`
	if err := b.NewPost("", path, title, body, "2025-06-23T13:17:30Z", nil, nil); err != nil {
		panic(err)
	}
}
