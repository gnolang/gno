// microblog is a website with shortform posts from users.
// the API is exteremly simple - "AddPost" takes markdown and
// adds it to the users site. The user is determined by
// the address and your microblog lives at
// /r/demo/microblog:<YOUR-ADDRESS>
//
// Getting started:
//
// (One-time) Add the microblog package:
//
//	gnokey maketx addpkg --pkgpath "gno.land/p/demo/microblog" --pkgdir "examples/gno.land/p/demo/microblog" \
//		--deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 <YOURKEY>
//
// (One-time) Add the microblog realm:
//
//	gnokey maketx addpkg --pkgpath "gno.land/r/demo/microblog" --pkgdir "examples/gno.land/r/demo/microblog" \
//		--deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 <YOURKEY>
//
// Add a microblog post:
//
//	gnokey maketx call --pkgpath "gno.land/r/demo/microblog" --func "NewPost" --args "hello, world" \
//		--gas-fee "1000000ugnot" --gas-wanted "2000000" --broadcast --chainid dev --remote localhost:26657 zkey
package microblog

import (
	"std"

	"gno.land/p/demo/microblog"
)

var (
	title  = "gno-based microblog"
	prefix = "/r/demo/microblog:"
	m      *microblog.Microblog
)

func init() {
	m = microblog.NewMicroblog(title, prefix)
}

// Render calls the microblog renderer
func Render(path string) string {
	return m.Render(path)
}

// NewPost takes a single argument (post markdown) and
// adds a post to the address of the caller.
func NewPost(text string) string {
	if m.NewPost(text) == "" {
		return "unable to add new post"
	}
	return "added new post"
}
