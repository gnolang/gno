package echo

import "gno.land/p/moul/md"

func Render(path string) string {
	// Fixed: escape attacker-controlled path text before inserting it in markdown.
	return "# Echo\n\n" + md.EscapeText(path)
}
