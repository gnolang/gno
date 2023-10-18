package present

func init() {
	path := "miami23"
	title := "Portal Loop Demo (Miami 2023)"
	body := `# Portal Loop Demo (Miami 2023)

- roadmap
- portal loop
- screenshot github project
`
	_ = b.NewPost(adminAddr, path, title, body, []string{"demo", "portal-loop", "miami"})
}
