package main

import (
	"std"
	"testing"

	"gno.land/p/lou/blog"
)

func main() {
	user := std.Address("g1testuser")
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))
	_, err := blog.NewBlog("", user)
	if err != nil {
		println("error:", err.Error())
	}
}

// Output:
// error: title cannot be empty
