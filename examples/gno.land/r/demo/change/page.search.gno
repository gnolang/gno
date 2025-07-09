package change

import (
	"gno.land/p/moul/md"
)

func renderSearchPage(y map[string]string) string {
	// content starts with topBar
	content := renderTopBar()

	// main page
	content += md.H1("Search a petition.") + br

	// TODO: use p/moul/md.Form to create the searchbar
	// 		 use p/pol/levenshtein to search something in searchbar with name
	//		 use p/demo/avl to search specific stuff

	return content
}