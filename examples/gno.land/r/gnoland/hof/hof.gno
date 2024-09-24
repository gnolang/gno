// Package hof represents the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"time"
)

var (
	permanent *Exhibition
	temporary *avl.Tree
	idCounter seqid.ID
)

type Exhibition struct {
	id seqid.ID

	title       string
	description string
	startTime   time.Time // given in RFC3339
	endTime     time.Time // end time of the event, given in RFC3339
	items       *avl.Tree
}

func init() {
	permanent = &Exhibition{
		id:          idCounter.Next(),
		title:       "Hall Of Fame",
		description: "gno.land's hall of fame!",
		startTime:   time.Now(),
		endTime:     time.Now(),
		items:       avl.NewTree(),
	}

	temporary = avl.NewTree()

}

func Register() {
	permanent

}

func RegisterTemp(id seqid.ID) {

}
