package echo

func Render(path string) string {
	// Vulnerable: raw path text can inject markdown into Render output.
	return "# Echo\n\n" + path
}
