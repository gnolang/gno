package main

import (
	"gno.land/p/demo/releases"
)

func main() {
	println("-----------")
	changelog := releases.NewChangelog("example")
	println(changelog.Render(""))

	println("-----------")
	changelog.NewRelease("v1", "r/blahblah", "* initial version")
	println(changelog.Render(""))

	println("-----------")
	changelog.NewRelease("v2", "r/blahblah2", "* various improvements\n* new shiny logo")
	println(changelog.Render(""))

	println("-----------")
	println(changelog.Latest().Render())
}

// Output:
// -----------
// # example
//
//
// -----------
// # example
//
// ## [example v1 (latest)](r/blahblah)
//
// * initial version
//
//
// -----------
// # example
//
// ## [example v2 (latest)](r/blahblah2)
//
// * various improvements
// * new shiny logo
//
// ## [example v1](r/blahblah)
//
// * initial version
//
//
// -----------
// ## [example v2 (latest)](r/blahblah2)
//
// * various improvements
// * new shiny logo
