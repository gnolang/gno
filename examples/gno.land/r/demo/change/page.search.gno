package change

import (
	"gno.land/p/moul/md"
)

func renderSearchPage(x string, y map[string]string) string {
	// content starts with topBar
	content := renderTopBar(x, y)

	// main page
	content += md.H1("Search a petition.") + br

	return content
}