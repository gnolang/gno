package change

import (
	"std"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"

	"strconv"
)

func SignPetition(_ realm, petitionId int) {
	// check if address is user
	addr := std.PreviousRealm().Address()
	user := users.ResolveAddress(addr)
	if user == nil {
		panic("expected address to be a connected user")
	}

	err := checkPetitionId(petitionId)
	if err != nil {
		panic(err)
	}

	s := &petitions[petitionId].signatures

	idToRemove := -1

	for i, e := range *s {
		if e == user {
			idToRemove = i
			break
		}
	}
	if idToRemove == -1 {
		*s = append(*s, user)
	} else {
		(*s)[idToRemove] = (*s)[len(*s)-1]
		*s = (*s)[:len(*s)-1]
	}
}

func CommentPetition(_ realm, petitionId int, content string) {
	// check if address is user
	addr := std.PreviousRealm().Address()
	user := users.ResolveAddress(addr)
	if user == nil {
		panic("expected address to be a connected user")
	}

	err := checkPetitionId(petitionId)
	if err != nil {
		panic(err)
	}

	c := &petitions[petitionId].comments
	*c = append(*c, comment{
		publisher: user,
		content:   content,
	})
}

func renderPetitionPage(params map[string]string) string {
	// the top bar
	content := renderTopBar()

	// check id
	idStr, okStr := params["id"]
	if !okStr {
		return content + "Missing id in parametters. Please try again."
	}
	id, err := checkStringPetitionId(idStr)
	if err != nil {
		return err.Error()
	}

	// render the content
	petition := petitions[id]
	content += md.H1(petition.title) + br
	content += "> By " + petition.publisher.RenderLink("") + br
	content += md.HorizontalRule() + br

	petitionContent := petition.description + br
	signContent := md.H1(strconv.Itoa(len(petition.signatures))) + md.H3("verified signatures")

	neededSigns := petition.target - len(petition.signatures)
	if neededSigns > 0 {
		signContent += "Only " + md.Bold(strconv.Itoa(neededSigns)) + " signatures left the reach the " + md.Bold(strconv.Itoa(petition.target)) + " signatures target !" + br
	} else {
		signContent += "Target of " + md.Bold(strconv.Itoa(petition.target)) + " signatures exceeded by " + md.Bold(strconv.Itoa(-neededSigns)) + " signatures !" + br
	}

	tx := txlink.Call("Sign", "id", strconv.Itoa(petition.id))
	signContent += md.H2(md.Link("Sign this petition now", tx))

	if len(petition.signatures) != 0 {
		signContent += md.H4("Last 5 signatures :")
		for i, e := range petition.signatures {
			if i >= 5 {
				break
			}
			signContent += " - " + e.RenderLink("") + br
		}
	}

	content += md.Columns([]string{petitionContent, signContent})

	content += md.H3("Comments (" + strconv.Itoa(len(petition.comments)) + "):")
	for _, e := range petition.comments {
		content += md.HorizontalRule() + br
		content += md.H4(e.publisher.RenderLink("")) + br
		content += e.content + br
	}

	return content
}
