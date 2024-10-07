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
		items       *avl.Tree // pkgPath > Item
		itemsSorted *avl.Tree // same data but sorted, storing pointers
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
		items:       avl.NewTree(),
		itemsSorted: avl.NewTree(),
	}

	owner = ownable.NewWithAddress(std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"))
	exhibition.Pausable = pausable.NewFromOwnable(owner)
}

// Register registers your realm to the Hall of Fame
// Should be called from within code
func Register() {
	if exhibition.IsPaused() {
		panic(ErrPaused.Error())
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
	i := &Item{
		id:       id,
		pkgpath:  pkgpath,
		deployer: std.GetOrigCaller(),
		blockNum: std.GetHeight(),
		upvote:   avl.NewTree(),
		downvote: avl.NewTree(),
	}

	exhibition.items.Set(pkgpath, i)
	exhibition.itemsSorted.Set(id.String(), i)
}

func Upvote(pkgpath string) {
	rawItem, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem.Error())
	}

	item := rawItem.(*Item)
	caller := std.PrevRealm().Addr().String()

	if item.upvote.Has(caller) {
		panic(ErrDoubleUpvote.Error())
	}

	item.upvote.Set(caller, struct{}{})
}

func Downvote(pkgpath string) {
	rawItem, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem.Error())
	}

	item := rawItem.(*Item)
	caller := std.PrevRealm().Addr().String()

	if item.downvote.Has(caller) {
		panic(ErrDoubleDownvote.Error())
	}

	item.downvote.Set(caller, struct{}{})
}

func Delete(pkgpath string) {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	i, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem.Error())
	}

	if _, removed := exhibition.itemsSorted.Remove(i.(*Item).id.String()); !removed {
		panic(ErrNoSuchItem.Error())
	}

	if _, removed := exhibition.items.Remove(pkgpath); !removed {
		panic(ErrNoSuchItem.Error())
	}
}
