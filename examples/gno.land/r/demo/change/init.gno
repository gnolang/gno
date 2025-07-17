package change

import (
	"gno.land/p/moul/md"
	"gno.land/r/sys/users"

	"std"
	"strconv"
	"strings"
)

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

type petition struct {
	publisher   *users.UserData
	title       string
	description string
	signatures  []*users.UserData
	target      int
	id          int
	comments    []comment
}

type comment struct {
	publisher *users.UserData
	content   string
}

func (p petition) LString() string {
	return p.title
}

func (p petition) MiniRender() string {
	content := md.HorizontalRule() + br
	content += md.H3(md.Link(p.title, here+":petition?id="+strconv.Itoa(p.id))) + br
	content += "by " + p.publisher.RenderLink("") + ", "
	content += md.Bold(strconv.Itoa(len(p.signatures))) + " / " + md.Bold(strconv.Itoa(p.target)) + " signatures." + br
	content += md.HorizontalRule() + br
	return content
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
