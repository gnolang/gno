package change

import (
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"

	"std"
)

func Submit(_ realm, title, description string) {
	// check if address is user
	addr := std.PreviousRealm().Address()
	user := users.ResolveAddress(addr)
	if user == nil {
		panic("expected address to be a connected user")
	}

	// push the stuff
	petitions = append(petitions, &petition{
		publisher:   *user,
		title:       title,
		description: description,
	})
}

func renderSubmitPage(params map[string]string) string {
	// content starts with topBar
	content := renderTopBar()

	// check if everything good
	var title, description, tx string
	title, ok := params["title"]
	if !ok {
		goto crappy
	}
	description, ok = params["description"]
	if !ok {
		goto crappy
	}

	// title
	content += md.H1("Submit your petition") + br
	content += "Here is what your petition will look like :" + br

	// show them
	content += md.HorizontalRule() + br
	content += md.H3(title) + br
	content += description + br
	content += md.HorizontalRule() + br

	// submit button
	tx = txlink.Call("Submit", "title", title, "description", description)
	content += md.H2(md.Link("Submit", tx)) + br
	content += "> [!NOTE]\n> You need to be a user to be able to submit." + br

	return content

crappy:
	return content + "Something went wrong with the parametters, please try again." + br
}
