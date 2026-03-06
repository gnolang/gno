package main

import (
	"testing"

	"gno.land/p/lou/blog"
)

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	_, err := blog.NewBlog("My Blog", address("he"))
	if err != nil {
		println("error:", err.Error())
	}
}

// Output:
// error: invalid caller address
