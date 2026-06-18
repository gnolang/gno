package echo

import "gno.land/p/moul/md"

func Render(path string) string {
	return "# Echo\n\n" + md.EscapeText(path)
}
