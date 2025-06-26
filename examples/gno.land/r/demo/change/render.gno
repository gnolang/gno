package change

import (
	"gno.land/p/moul/md"
)

const br = "\n\n"

func renderMainPage() string {
	var content string

	// main page
	content += md.H1("r/change") + br

	// petition button
	content += md.H2(md.Link("Start a petition", "")) + br

	content += `
<gno-form>
	<gno-input name="name" placeholder="Enter your name" />

	hello

	<gno-input name="email" placeholder="Enter your email" />
</gno-form>
	`

	return content
}

func Render(_ string) string {
	return renderMainPage()
}
