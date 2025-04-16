package home

import (
	"gno.land/p/moul/md"
)

func Render(path string) (rend string) {
	rend += md.H1("Leo's home page")
	rend += md.H2("Who am I ?")
	rend += "We'll... I have litterally no idea but I like reverse eng, cyber, low-level and when i have to manipulate those bytes by myself.\n\n"
	rend += "(never made a portfolio I have no clue what the hell i have to put here)\n\n"

	rend += md.H2("Wanna Play ?")
	rend += md.H3(md.Link("Pokémon", "pokemon"))

	return
}
