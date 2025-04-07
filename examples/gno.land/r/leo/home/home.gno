package home

import (
	"gno.land/p/leo/renderer"
	"gno.land/r/leon/hof"
)

func init() {
	hof.Register("Leo's", "Come play some games")
}

func Render(path string) (rend string) {
	rend += renderer.Header(1, "Leo's home page")
	rend += renderer.Header(2, "Who am I ?")
	rend += "We'll... I have litterally no idea but I like reverse eng, cyber, low-level and when i have to manipulate those bytes by myself.\n\n"
	rend += "(never made a portfolio I have no clue what the hell i have to put here)\n\n"

	rend += renderer.Header(2, "Wanna Play ?")
	rend += renderer.Header(3, renderer.Link("Pokemon", "home/pokemon"))

	return
}
