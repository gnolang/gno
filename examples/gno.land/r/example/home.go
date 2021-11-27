package example

import (
	"gno.land/p/groups"
	// TODO: make this manual switcharoo unnecessary.
	// "github.com/gnolang/gno/examples/gno.land/p/groups"
)

var group *groups.Group

func init() {
	group = &groups.Group{Name: "First Group"}
}

func AddPost(title string, body string) {
	group.AddPost(title, body)
}

func Render() string {
	return group.String()
}
