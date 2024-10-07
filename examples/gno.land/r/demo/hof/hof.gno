// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"
	"gno.land/p/demo/seqid"
)

var (
	exhibition *Exhibition
	owner      *ownable.Ownable
)

type (
	Exhibition struct {
		itemCounter seqid.ID
		description string
		items       *avl.Tree // Item id > Item
		*pausable.Pausable
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
	}

	owner = ownable.NewWithAddress(std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"))
	exhibition.Pausable = pausable.NewFromOwnable(owner)
}

// Register registers your realm to the Hall of Fame
// Should be called from within code
func Register() {
	if exhibition.IsPaused() {
		panic("hof: registering is paused")
	}

	submission := std.PrevRealm()
	pkgpath := submission.PkgPath()

	// Must not yet exist and must be called from code
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
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	if _, removed := exhibition.items.Remove(itemID.String()); !removed {
		return ErrNoSuchItem
	}

	return nil
}
