// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/lou/blog"
)

func main(cur realm) {
	_, err := blog.NewBlog("My Blog", address("he"), cur)
	if err != nil {
		println("error:", err.Error())
	}
}

// Output:
// error: invalid caller address
