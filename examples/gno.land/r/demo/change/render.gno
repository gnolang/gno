package change

import (
	"gno.land/p/moul/md"
)

const br = "\n\n"

func renderMainPage() string {
	var content string

	// main page
	content += md.H1("r/change")

	return content
}

func Render(_ string) string {
	return renderMainPage()
}
