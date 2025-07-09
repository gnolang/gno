package change

import (
	"gno.land/p/moul/md"

	"strconv"
)

func renderPetitionPage(params map[string]string) string {
	// the top bar
	content := renderTopBar()

	// check id
	idStr, okStr := params["id"]
	if !okStr {
		return content + "Missing id in parametters. Please try again."
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return content + err.Error() + ". Please try again."
	}
	if id < 0 || id > len(petitions)-1 {
		return "Invalid petition id range (too big or too small). Please try again."
	}

	// render the content
	petition := petitions[id]
	content += md.H1(petition.title) + br
	content += petition.description + br
	content += "> By " + petition.publisher.RenderLink("")

	return content
}