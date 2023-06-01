package main

import (
	"gno.land/r/demo/releases_example"
)

func main() {
	println("-----------")
	print(releases_example.Render(""))

	println("-----------")
	print(releases_example.Render("v1"))

	println("-----------")
	print(releases_example.Render("v42"))
}

// Output:
// -----------
// # example_app
//
// ## [example_app v2 (latest)](r/demo/examples_example_v2)
//
// various improvements
//
// ## [example_app v1](r/demo/examples_example_v1)
//
// initial release
//
// -----------
// ## [example_app v1](r/demo/examples_example_v1)
//
// initial release
//
// -----------
// no such release
