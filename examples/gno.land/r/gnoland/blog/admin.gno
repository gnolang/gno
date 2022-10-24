package gnoblog

import (
	"std"
	"strings"
)

var adminAddr std.Address

func init() {
	adminAddr = std.GetOrigCaller()
}

func AdminNewPost(slug, title, body, tags string) {
	assertIsAdmin()
	b.NewPost(slug, title, body, strings.Split(tags, ","))
}

func AdminUpdateAdminAddr(newAddr std.Address) {
	assertIsAdmin()
	adminAddr = newAddr
}

func assertIsAdmin() {
	if std.GetOrigCaller() != adminAddr {
		panic("access restricted.")
	}
}
