package main

import (
	"gno.land/p/demo/releases"
)

func main() {
	println("-----------")
	changelog := releases.NewChangelog("example")
	print(changelog.Render(""))

	println("-----------")
	changelog.NewRelease("v1", "r/blahblah", "* initial version")
	print(changelog.Render(""))

	println("-----------")
	changelog.NewRelease("v2", "r/blahblah2", "* various improvements\n* new shiny logo")
	print(changelog.Render(""))

	println("-----------")
	print(changelog.Latest().Render())
}

// Output:
// -----------
// # example
//
// -----------
// # example
//
// ## [example v1 (latest)](r/blahblah)
//
// * initial version
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
// -----------
// ## [example v2 (latest)](r/blahblah2)
//
// * various improvements
// * new shiny logo
//
