package main

import (
	"gno.land/r/releases-example"
)

func main() {
	println("-----------")
	println(releases_example.Render(""))

	println("-----------")
	println(releases_example.Render("v1"))

	println("-----------")
	println(releases_example.Render("v42"))
}

// Output:
// -----------
// # example-app
//
// ## [example-app v2 (latest)](r/examples-example-v2)
//
// various improvements
//
// ## [example-app v1](r/examples-example-v1)
//
// initial release
//
//
// -----------
// ## [example-app v1](r/examples-example-v1)
//
// initial release
//
//
// -----------
// no such release
