package awesomegno

import (
	"std"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

type Category struct {
	ID          seqid.ID
	Name        string
	Description string
	CreatedAt   time.Time
	items       *avl.Tree
}

type Item struct {
	ID          seqid.ID
	Title       string
	Description string
	Authors     []std.Address
	URL         string
	CreatedAt   time.Time
	Categories  []seqid.ID
}
