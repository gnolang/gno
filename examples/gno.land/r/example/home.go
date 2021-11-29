package example

import (
	"std"

	"gno.land/p/groups"
	// TODO: make this manual switcharoo unnecessary.
	// "github.com/gnolang/gno/examples/gno.land/p/groups"
)

var group *groups.Group

func init() {
	group = &groups.Group{Name: "First Group"}
}

func AddPost(title string, body string) uint64 {
	// TODO: consider making this a function tag/decorator.
	if !std.IsOriginCall() {
		panic("AddPost is public facing")
	}
	ctx := std.GetContext()
	post := group.AddPost(ctx.Caller, title, body)
	return post.ID
}

func Render() string {
	return group.String()
}
