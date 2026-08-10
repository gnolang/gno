// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/lou/blog"
)

func main(cur realm) {
	user := address("g1testuser")
	_, err := blog.NewBlog(0, cur, "", user)
	if err != nil {
		println("error:", err.Error())
	}
}

// Output:
// error: title cannot be empty
