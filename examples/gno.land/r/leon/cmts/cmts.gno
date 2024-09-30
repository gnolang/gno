// Package cmts a simple, pluggable, yet configurable comment functionality set
package cmts

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

type CommentSet struct {
	owner *ownable.Ownable
	set   *avl.Tree
}

type Comment struct {
	text      string
	author    std.Address
	up, down  int
	replies   *avl.Tree
	deletable bool
}

func NewCommentSet() *CommentSet {
	return &CommentSet{
		owner: ownable.New(),
		set:   avl.NewTree(),
	}
}
