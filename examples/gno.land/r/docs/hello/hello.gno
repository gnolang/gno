// Package hello demonstrates basic usage of Render().
// Try adding `:World` at the end of the URL, like `.../hello:World`.
//
// Render() is optional: a realm without it still has a gnoweb page, showing
// its file listing instead of rendered markdown.
// When present, it must have the exact signature: func Render(path string) string.
package hello

import "gno.land/p/nt/markdown/sanitize/v0"

// Render outputs a greeting. It customizes the message based on the provided path.
// path is attacker-controlled (it comes from the URL), so it is sanitized before
// being embedded in markdown output.
func Render(path string) string {
	if path == "" {
		return "# Hello, 世界！"
	}
	return "# Hello, " + sanitize.InlineText(path) + "!"
}
