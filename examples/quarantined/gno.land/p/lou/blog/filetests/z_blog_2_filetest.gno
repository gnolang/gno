// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/lou/blog"
)

func main(cur realm) {
	user := address("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n")
	testBlog, err := blog.NewBlog(0, cur, "test blog", user)
	if err != nil {
		println(err.Error())
	}
	println(testBlog.Render(""))
}

// Output:
// # test blog
//
// [⊞ grid](/r/test:?mode=grid) | [≔ list](/r/test:?mode=list) | [⧖ relative](/r/test:?time=short) | [↕ alphabetical \(A\-Z\)](/r/test:?order=asc&sort=alpha) | [↕ recent](/r/test:?order=asc&sort=recent) | [↕ last updated](/r/test:?order=asc&sort=update) | [past year](/r/test:?end=2009-02-13&start=2008-02-13), [this year](/r/test:?end=2009-02-13&start=2009-01-01), [last 30 days](/r/test:?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/test)
//
// No posts found.
