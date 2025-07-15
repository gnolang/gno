package change

import (
	"std"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"

	"strconv"
)

func Sign(_ realm, id int) {
	// check if address is user
	addr := std.PreviousRealm().Address()
	user := users.ResolveAddress(addr)
	if user == nil {
		panic("expected address to be a connected user")
	}

	err := checkPetitionId(id)
	if err != nil {
		panic(err)
	}

	idToRemove := -1

	for i, e := range petitions[id].signatures {
		if e == user {
			idToRemove = i
			break
		}
	}
	if idToRemove == -1 {
		petitions[id].signatures = append(petitions[id].signatures, user)
	} else {
		petitions[id].signatures[idToRemove] = petitions[id].signatures[len(petitions[id].signatures)-1]
		petitions[id].signatures = petitions[id].signatures[:len(petitions[id].signatures)-1]
	}
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
		signContent += md.H4("Already signed by :")
		for _, e := range petition.signatures {
			signContent += " - " + e.RenderLink("") + br
		}
	}

	content += md.Columns([]string{petitionContent, signContent})

	return content
}
