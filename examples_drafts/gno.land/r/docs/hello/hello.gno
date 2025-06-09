// Package hello_world demonstrates basic usage of Render().
// Try adding `:World` at the end of the URL, like `.../hello:World`.
package hello

// Render outputs a greeting. It customizes the message based on the provided path.
func Render(path string) string {
	if path == "" {
		return "# Hello, 世界！"
	}
	return "# Hello, " + path + "!"
}
