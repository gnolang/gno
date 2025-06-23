package gnopages

func init() {
	path := "partners"
	title := "Partnerships"
	body := `# Useful links

Link1

Link2

`
	_ = b.NewPost("", path, title, body, "2025-06-23T13:17:27Z", nil, nil)
}
