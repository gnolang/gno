// http://localhost:8888/r/pol/tofix/multiparams$help&func=Toto&description=this+is+a+description&title=this+is+a+title

package multiparams

import (
	"gno.land/p/moul/txlink"
)

var show string = "nothing for now"

func Toto(title string, description string) {
	show = "title: " + title + ", description: " + description
}

func Render(_ string) string {
	lnk := txlink.Call("Toto", "title", "this is a title", "description", "this is a description")
	// lnk := txlink.Call("Toto", "description", "this is a description", "title", "this is a title")
	return show + " / [click here](" + lnk + ")\n"
}
