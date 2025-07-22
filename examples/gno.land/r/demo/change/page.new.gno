package change

import (
	"gno.land/p/moul/md"
)

func renderNewPage(_ map[string]string) string {
	// content starts with topBar
	content := renderTopBar()

	// title
	content += md.H1("Create a new petition") + br

	// create the title
	// TODO: replace this by p/moud/md.Form
	content += `<gno-form path="submit">` + br
	content += `  <gno-input name="title" placeholder="Title of the petition" />` + br
	content += `  <gno-input name="description" placeholder="Describe the cause" />` + br
	content += `  <gno-input type="number" name="target" placeholder="Target number of signatures" />` + br
	content += `  <gno-input name="location" placeholder="Location (leave empty if global)" />` + br
	content += "</gno-form>" + br

	return content
}
