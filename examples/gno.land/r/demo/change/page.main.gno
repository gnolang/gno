package change

import (
	"gno.land/p/moul/md"
)

func renderMainPage(_ map[string]string) string {
	// content starts with topBar
	content := renderTopBar()

	// main page
	content += md.H1("Change starts now.") + br

	// petition button
	content += md.H2(md.Link("Start a petition", here+":new")) + br + md.HorizontalRule() + br

	// lil message
	content += md.H2("How do we help you make a difference") + br
	one := md.H3("1. Start a petition") + br + "In just minutes, create a quality campaign with our expert advice to make a difference."
	two := md.H3("2. Gather support") + br + "Reach people who care about your cause with our distribution tools and our extensive network of partners."
	three := md.H3("3. Reaching out to those in power") + br + "Influence officials in government, education, and other institutions with our proven process."
	content += md.Columns([]string{one, two, three}, false) + md.HorizontalRule() + br

	// petition selection
	content += md.H2("Support the causes you care about") + br
	content += "Find petitions that matter to you and make your voice heard to make a difference."

	// TODO: petition selection, with subjects like on https://www.change.org/

	return content
}
