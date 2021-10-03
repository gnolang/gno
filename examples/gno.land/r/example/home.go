package example

import (
	"gno.land/p/dom"
	// TODO: make this manual switcharoo unnecessary.
	// "github.com/gnolang/gno/examples/gno.land/p/dom"
)

var gPlot *dom.Plot

func init() {
	gPlot = &dom.Plot{Name: "my plot"}
}

func AddPost(title string, body string) {
	gPlot.AddPost(title, body)
}

func RenderPlot() string {
	return gPlot.String()
}
