package change

import (
	"gno.land/p/moul/md"
)

func renderSearchResult(params map[string]string) string {
	// get levenshteinable
	xs := make([]Levenshteinable, len(petitions))
	for k, v := range petitions {
		xs[k] = v
	}

	// check best
	q := params["q"]
	content := md.H1("Results for " + md.InlineCode(q) + ":")
	best := SeveralPickFromString(q, xs, 10)
	bestStrs := make([]string, len(best))
	for k, v := range best {
		bestStrs[k] = v.(*petition).MiniRender()
	}

	// render
	content += md.ColumnsN(bestStrs, 3, true) + br
	return content
}

func renderSearchPage(params map[string]string) string {
	// content starts with topBar
	content := renderTopBar()

	// check if result
	if _, ok := params["q"]; ok {
		return content + renderSearchResult(params)
	}

	// main page
	content += md.H1("Search a petition.") + br

	// searchbar
	// TODO: use p/moul/md.Form to create the searchbar
	// 		 use p/pol/levenshtein to search something in searchbar with name
	//		 use p/demo/avl to search specific stuff
	content += `<gno-form path="search">` + br
	content += `  <gno-input name="q" placeholder="Search" />` + br
	content += "</gno-form>" + br

	return content
}