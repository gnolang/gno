// PKGPATH: gno.land/r/example
package example

import (
	"github.com/gnolang/gno/_test/dom"
)

var gPlot *dom.Plot

func init() {
	gPlot = &dom.Plot{Name: "First Plot"}
}

func main() {
	gPlot.AddPost("TEST_TITLE", "TEST_BODY")
	println("done")
}

// Output:
// done

// Realm:
// XXX
