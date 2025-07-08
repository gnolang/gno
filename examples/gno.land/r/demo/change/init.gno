package change

import (
	"gno.land/p/moul/md"

	"strings"
	"std"
)

type user struct {
	addr address
}

var (
	here = std.CurrentRealm().PkgPath()[len(std.ChainDomain()):]
	users = []user{}
	bigSeparator = strings.Repeat("‎ ", 10)

	pages = map[string]func(string, map[string]string) string {
		"": renderMainPage,
		"search": renderSearchPage,
	}
)

const br = "\n\n"

func renderTopBar(_ string, params map[string]string) string {
	var content string

	// home
	content += md.Link("r/change", here + ":new") + bigSeparator

	// search
	content += md.Link("Search", here + ":search") + bigSeparator

	// start
	content += md.Link("Start a petition", here + ":new") + bigSeparator

	return content + br
}

func Render(query string) string {
	x, y := ParseQuerry(query)
	if page, ok := pages[x]; ok {
		return page(x, y)
	}
	return renderMainPage(x, y)
}
