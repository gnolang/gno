package main

import (
	"std"
	"testing"

	"gno.land/p/lou/blog"
)

func main() {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))
	_, err := blog.NewBlog("My Blog", std.Address("he"))
	if err != nil {
		println("error:", err.Error())
	}
}

// Output:
// error: invalid caller address
