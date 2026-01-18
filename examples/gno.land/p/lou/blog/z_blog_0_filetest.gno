package main

import (
	"testing"

	"gno.land/p/lou/blog"
)

func main() {
	user := address("g1testuser")
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	_, err := blog.NewBlog("", user)
	if err != nil {
		println("error:", err.Error())
	}
}

// Output:
// error: title cannot be empty
