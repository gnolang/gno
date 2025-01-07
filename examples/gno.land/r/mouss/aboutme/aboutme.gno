package aboutme

import (
	"gno.land/p/demo/mux"
)

var abtMe string

func init() {
	abtMe = "Hi, I'm Mustapha, a contributor to gno.land."
}
func renderAboutMe(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(abtMe)
}
