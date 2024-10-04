// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

var (
	exhibCounter seqid.ID
	exhibitions  *avl.Tree // id > *Exhibition
)

type (
	Exhibition struct {
		id          seqid.ID // fixed per exhibition
		itemCounter seqid.ID
		title       string
		description string
		startTime   *time.Time // given in RFC3339
		endTime     *time.Time // end time of the event, given in RFC3339
		items       *avl.Tree  // Item id > Item
	}

	Item struct {
		id       seqid.ID
		exhibID  seqid.ID
		pkgpath  string
		deployer std.Address
		blockNum int64
		upvote   *avl.Tree // std.Addr > struct{}{}
		downvote *avl.Tree // std.Addr > struct{}{}
	}
)

func init() {
	exhibitions = avl.NewTree()

	exhibitions.Set(seqid.ID(0).String(),
		&Exhibition{ // permanent exhibition, ID=0
			id:    seqid.ID(0),
			items: avl.NewTree(),
			title: "Permanent Exhibition",
		})
}

// CreateExhibition creates a temporary exhibition to be displayed while it's active
func CreateExhibition(title, description string, start, end *time.Time) error {
	// todo add admin checks

	if strings.TrimSpace(title) == "" {
		return ErrInvalidTitle
	}

	id := exhibCounter.Next()
	e := &Exhibition{
		id:          id,
		title:       title,
		description: description,
		items:       avl.NewTree(),
		startTime:   start,
		endTime:     end,
	}

	exhibitions.Set(id.String(), e)

	return nil
}

// Register registers to the permanent exhibition
func Register() {
	RegisterFor(seqid.ID(0))
}

// RegisterFor registers your realm for a specific exhibition
func RegisterFor(exhibID seqid.ID) {
	submission := std.PrevRealm()
	pkgpath := submission.PkgPath()

	raw, ok := exhibitions.Get(exhibID.String())
	if !ok {
		panic(ErrNoSuchExhibition.Error())
	}

	e := raw.(*Exhibition)

	// Must not already exist and must be called from code
	if submission.IsUser() {
		panic(ErrNonCodeCall.Error())
	}

	if e.items.Has(pkgpath) {
		panic(ErrAlreadyExists.Error())
	}

	id := e.itemCounter.Next()
	e.items.Set(
		id.String(),
		&Item{
			id:       id,
			exhibID:  e.id,
			pkgpath:  pkgpath,
			deployer: std.GetOrigCaller(),
			blockNum: std.GetHeight(),
			upvote:   avl.NewTree(),
			downvote: avl.NewTree(),
		})
}

func Upvote(exhibID, itemID seqid.ID) error {
	raw, ok := exhibitions.Get(exhibID.String())
	if !ok {
		return ErrNoSuchExhibition
	}

	e := raw.(*Exhibition)

	rawItem, ok := e.items.Get(itemID.String())
	if !ok {
		return ErrNoSuchItem
	}

	item := rawItem.(*Item)
	caller := std.PrevRealm().Addr().String()

	if item.upvote.Has(caller) {
		return ErrDoubleUpvote
	}

	item.upvote.Set(caller, struct{}{})
	return nil
}

func Downvote(exhibID, itemID seqid.ID) error {
	raw, ok := exhibitions.Get(exhibID.String())
	if !ok {
		return ErrNoSuchExhibition
	}

	e := raw.(*Exhibition)

	rawItem, ok := e.items.Get(itemID.String())
	if !ok {
		return ErrNoSuchItem
	}

	item := rawItem.(*Item)
	caller := std.PrevRealm().Addr().String()

	if item.downvote.Has(caller) {
		return ErrDoubleDownvote
	}

	item.downvote.Set(caller, struct{}{})
	return nil
}

func Delete(exhibID, itemID seqid.ID) error {
	raw, ok := exhibitions.Get(exhibID.String())
	if !ok {
		return ErrNoSuchExhibition
	}

	if _, removed := raw.(*Exhibition).items.Remove(itemID.String()); !removed {
		return ErrNoSuchItem
	}

	return nil
}
