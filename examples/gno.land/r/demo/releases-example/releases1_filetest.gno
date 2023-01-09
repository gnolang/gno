package main

import (
	"gno.land/r/demo/releases-example"
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
// # example-app
//
// ## [example-app v2 (latest)](r/demo/examples-example-v2)
//
// various improvements
//
// ## [example-app v1](r/demo/examples-example-v1)
//
// initial release
//
// -----------
// ## [example-app v1](r/demo/examples-example-v1)
//
// initial release
//
// -----------
// no such release
