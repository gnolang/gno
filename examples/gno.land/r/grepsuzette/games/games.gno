package games

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

func Render(path string) string {
	return "#### r/grepsuzette/games\n\n" +
		"A collection of small games written in the Gno language:\n\n" +
		renderGames() + "\n\n"
}

func renderGames() string {
	var b strings.Builder
	for _, o := range []struct {
		title, path string
		authors     []string
		desc        string
	}{
		{"tictactoe", "games/tictactoe", []string{"grepsuzette", "moul"}, "CPU vs Human tictactoe, no need for a wallet"},
		{"underwater minesweeper", "games/minesweeper", []string{"grepsuzette"}, "A classic productivity tool"},
	} {
		b.WriteString(ufmt.Sprintf(
			"* [%s](%s): %s by %s\n",
			o.title, o.path,
			o.desc,
			RenderAuthors(o.authors, "@"),
		))
	}
	return b.String()
}

// ["a", "b", "c"] -> "a, b and c"
// Typical prefix is "@": "a", "b", "c" -> "@a, @b and @c"
func RenderAuthors(authors []string, optionalPrefix ...string) string {
	a := []string{}
	prefix := ""
	if len(optionalPrefix) > 0 {
		prefix = optionalPrefix[0]
	}
	for _, author := range authors {
		a = append(a, prefix+author)
	}
	switch len(a) {
	case 0:
		return "*?*"
	case 1:
		return a[0]
	default:
		return strings.Join(a[0:len(a)-1], ", ") + " and " + a[len(a)-1]
	}
}
