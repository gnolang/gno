package optional_render

func Info() string {
	return `Having a Render() function in your realm is optional!
If you do decide to have a Render() function, it must have the following signature:
func Render(path string) string { ... }`
}
