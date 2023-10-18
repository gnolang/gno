package present

func init() {
	path := "miami23"
	title := "Portal Loop Demo (Miami 2023)"
	body := `# Portal Loop Demo (Miami 2023)

- subject A
- subject B
- subject C
`
	_ = b.NewPost("", path, title, body, nil)
}
