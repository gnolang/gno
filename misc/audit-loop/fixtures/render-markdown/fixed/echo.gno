package echo

import (
	"html"

	"gno.land/p/moul/md"
)

func Render(path string) string {
	return "# Echo\n\n" + md.EscapeText(html.EscapeString(path))
}
