package deep

func Render(path string) string {
	if path == "" {
		return "it works!"
	} else {
		return "hi " + path
	}
}
