package change

import (
	"strconv"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"

	"std"
)

func Submit(cur realm, title, description, target, location string) {
	// check if address is user
	addr := std.PreviousRealm().Address()
	user := users.ResolveAddress(addr)
	if user == nil {
		panic("expected address to be a connected user")
	}

	targetId, err := strconv.Atoi(target)
	if !isStringPositiveInt(target) || err != nil {
		panic("target not a positive int")
	}

	// push the stuff
	petitions = append(petitions, &petition{
		publisher:   user,
		title:       title,
		description: description,
		signatures:  []*users.UserData{},
		target:      targetId,
		id:          len(petitions),
		location:    location,
		date:        time.Now(),
	})
}

func renderSubmitPage(params map[string]string) string {
	// content starts with topBar
	content := renderTopBar()

	// check if everything good
	var title, description, target, location, tx string
	title, ok := params["title"]
	if !ok {
		goto crappy
	}
	description, ok = params["description"]
	if !ok {
		goto crappy
	}
	target, ok = params["target"]
	if !ok || !isStringPositiveInt(target) {
		goto crappy
	}
	location, ok = params["location"]
	if !ok {
		goto crappy
	}

	// title
	content += md.H1("Submit your petition") + br
	content += "Here is what your petition will look like :" + br

	// show them
	content += md.HorizontalRule() + br
	content += md.H2(title) + br
	content += description + br
	content += md.HorizontalRule() + br

	content += md.Bold(params["target"]) + " signatures target" + br

	// submit button
	tx = txlink.Call("Submit", "title", title, "description", description, "target", target, "location", location)
	content += md.H2(md.Link("Submit", tx)) + br
	content += "> [!NOTE]\n> You need to be a user to be able to submit (" + md.Link("gno.land/r/gnoland/users", "/r/gnoland/users") + ")." + br

	return content

crappy:
	return content + "Something went wrong with the parametters, please try again." + br
}
