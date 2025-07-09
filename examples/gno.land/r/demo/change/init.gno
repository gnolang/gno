package change

import (
	"gno.land/p/moul/md"
	"gno.land/r/sys/users"

	"std"
	"strings"
)

type petition struct {
	publisher   users.UserData
	title       string
	description string
}

var (
	here         = std.CurrentRealm().PkgPath()[len(std.ChainDomain()):]
	petitions    = []*petition{}
	bigSeparator = strings.Repeat("‎ ", 10)

	pagesRouter = map[string]func(map[string]string) string{
		"":         renderMainPage,
		"search":   renderSearchPage,
		"new":      renderNewPage,
		"submit":   renderSubmitPage,
		"petition": renderPetitionPage,
	}
)

const br = "\n\n"

func renderTopBar() string {
	var content string

	// home
	content += md.Link("r/change", here) + bigSeparator

	// search
	content += md.Link("Search", here+":search") + bigSeparator

	// start
	content += md.Link("Start a petition", here+":new") + bigSeparator

	return content + br
}

func Render(query string) string {
	x, y := ParseQuerry(query)
	if page, ok := pagesRouter[x]; ok {
		return page(y)
	}
	return renderMainPage(y)
}
