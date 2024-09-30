// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"std"
	"strings"
	"time"
)

var (
	exhibCounter seqid.ID
	permanent    *Exhibition
	temporary    *avl.Tree // id > *Exhibition
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
		pkgpath  string
		deployer std.Address
		blockNum int64
		upvote   *avl.Tree // std.Addr > struct{}{}
		downvote *avl.Tree // std.Addr > struct{}{}
	}
)

func init() {
	permanent = &Exhibition{
		id:    exhibCounter.Next(),
		title: "Permanent Exhibition",
		items: avl.NewTree(),
	}

	temporary = avl.NewTree()
}

// Register registers your realm to the permanent exhibition
func Register() {
	submission := std.PrevRealm()
	pkgpath := submission.PkgPath()

	// must not exist and must be called from code
	if permanent.items.Has(pkgpath) || submission.IsUser() {
		panic("item exists or not code call")
	}

	id := permanent.itemCounter.Next()
	permanent.items.Set(
		id.String(),
		&Item{
			id:       id,
			pkgpath:  pkgpath,
			deployer: std.GetOrigCaller(),
			blockNum: std.GetHeight(),
			upvote:   avl.NewTree(),
			downvote: avl.NewTree(),
		})
}

// CreateTemporaryExhibition creates a temporary exhibition to be displayed while it's active
func CreateTemporaryExhibition(title, description string, start, end *time.Time) error {
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

	temporary.Set(id.String(), e)

	return nil
}

// RegisterFor registers your realm for a specific exhibition
//func RegisterFor(exhibID string) {
//	submission := std.PrevRealm()
//	pkgpath := submission.PkgPath()
//
//	raw, ok := temporary
//
//	// must not exist and must be called from code
//	if permanent.items.Has(pkgpath) || submission.IsUser() {
//		panic("item exists or not code call")
//	}
//
//	id := permanent.itemCounter.Next()
//	permanent.items.Set(
//		id.String(),
//		&Item{
//			id:       id,
//			pkgpath:  pkgpath,
//			deployer: std.GetOrigCaller(),
//			blockNum: std.GetHeight(),
//			upvote:   avl.NewTree(),
//			downvote: avl.NewTree(),
//		})
//}

func Upvote(exhibID, itemID seqid.ID) error {
	if exhibID == seqid.ID(0) { // permanent exhibition
		raw, ok := permanent.items.Get(itemID.String())
		if !ok {
			return ErrNoSuchItem
		}

		item := raw.(*Item)
		caller := std.PrevRealm().Addr().String()

		if item.upvote.Has(caller) {
			return ErrDoubleUpvote
		}

		item.upvote.Set(caller, struct{}{})
		return nil
	}

	// todo add temp exh
	return nil
}

func Downvote(exhibID, itemID seqid.ID) error {
	if exhibID == seqid.ID(0) { // permanent exhibition
		raw, ok := permanent.items.Get(itemID.String())
		if !ok {
			return ErrNoSuchItem
		}

		item := raw.(*Item)
		caller := std.PrevRealm().Addr().String()

		if item.downvote.Has(caller) {
			return ErrDoubleDownvote
		}

		item.downvote.Set(caller, struct{}{})
		return nil
	}

	// todo add temp exh
	return nil
}
