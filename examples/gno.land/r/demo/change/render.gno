package change

import (
	"gno.land/p/moul/md"
	"strconv"
	"strings"
)

type user struct {
	addr address
}

var (
	users = []user{}
	bigSeparator = strings.Repeat("‎ ", 10)
)

const br = "\n\n"

func renderTopBar(_ string, params map[string]string) string {
	var content string

	// home
	content += md.Link("r/change", "") + bigSeparator

	// search
	content += md.Link("Search", "") + bigSeparator

	// login
	content += md.Link()

	return content + br
}

func renderMainPage(x string, y map[string]string) string {
	// content starts with topBar
	content := renderTopBar(x, y)

	// main page
	content += md.H1("Change starts now.") + br

	// counter only if 10+ people joined, because else its weird
	userLen := len(users)
	if userLen >= 10 {
		content += "Joins " + md.Bold(strconv.Itoa(userLen)) + " users." + br
	}

	// petition button
	content += md.H2(md.Link("Start a petition", "")) + br

	return content
}

func Render(query string) string {
	x, y := ParseQuerry(query)
	return renderMainPage(x, y)
}
