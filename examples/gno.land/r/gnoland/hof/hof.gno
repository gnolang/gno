// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

var exhibition *Exhibition

type (
	Exhibition struct {
		itemCounter seqid.ID
		title       string
		description string
		items       *avl.Tree // Item id > Item
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
	exhibition = &Exhibition{
		items: avl.NewTree(),
		title: "Hall of Fame",
	}
}

// Register registers your realm for a specific exhibition
func Register() {
	submission := std.PrevRealm()
	pkgpath := submission.PkgPath()

	// Must not already exist and must be called from code
	if submission.IsUser() {
		panic(ErrNonCodeCall.Error())
	}

	if exhibition.items.Has(pkgpath) {
		panic(ErrAlreadyExists.Error())
	}

	id := exhibition.itemCounter.Next()
	exhibition.items.Set(
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

func Upvote(itemID seqid.ID) error {
	rawItem, ok := exhibition.items.Get(itemID.String())
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

func Downvote(itemID seqid.ID) error {

	rawItem, ok := exhibition.items.Get(itemID.String())
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

func Delete(itemID seqid.ID) error {
	if _, removed := exhibition.items.Remove(itemID.String()); !removed {
		return ErrNoSuchItem
	}

	return nil
}
