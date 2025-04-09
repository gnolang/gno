package awesomegno

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

type Category struct {
	ID          seqid.ID
	Name        string
	Description string
	CreatedAt   time.Time
	dapps       *avl.Tree
}

type Dapp struct {
	ID          seqid.ID
	Title       string
	Description string
	Author      std.Address
	URL         string
	CreatedAt   time.Time
	categories  *avl.Tree
}
