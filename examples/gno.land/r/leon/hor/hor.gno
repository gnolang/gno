// Package hor is the hall of realms.
// The Hall of Realms is an exhibition that holds items. Users can add their realms to the Hall of Realms by
// importing the Hall of Realms package and calling hor.Register() from their init function.
package hor

import (
	"chain"
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/pausable"
	"gno.land/p/nt/seqid"
	"gno.land/r/leon/config"
)

const (
	maxTitleLength       = 30
	maxDescriptionLength = 50
)

var (
	exhibition *Exhibition

	// Safe objects
	Ownable  *ownable.Ownable
	Pausable *pausable.Pausable
)

type (
	Exhibition struct {
		itemCounter            seqid.ID
		description            string
		items                  *avl.Tree // pkgPath > *Item
		itemsSortedByCreation  *avl.Tree // same data but sorted by creation time
		itemsSortedByUpvotes   *avl.Tree // same data but sorted by upvotes
		itemsSortedByDownvotes *avl.Tree // same data but sorted by downvotes
	}

	Item struct {
		id          seqid.ID
		title       string
		description string
		pkgpath     string
		blockNum    int64
		upvote      *addrset.Set
		downvote    *addrset.Set
	}
)

func init() {
	exhibition = &Exhibition{
		items:                  avl.NewTree(),
		itemsSortedByCreation:  avl.NewTree(),
		itemsSortedByUpvotes:   avl.NewTree(),
		itemsSortedByDownvotes: avl.NewTree(),
	}

	Ownable = ownable.NewWithAddressByPrevious(config.OwnableMain.Owner()) // OrigSendOwnable?
	Pausable = pausable.NewFromOwnable(Ownable)
}

// Register registers your realm to the Hall of Fame
// Should be called from within code
func Register(cur realm, title, description string) {
	if Pausable.IsPaused() {
		return
	}

	submission := runtime.PreviousRealm()
	pkgpath := submission.PkgPath()

	// Must be called from code
	if submission.IsUser() {
		return
	}

	// Must not yet exist
	if exhibition.items.Has(pkgpath) {
		return
	}

	// Title must be between 1 maxTitleLength long
	if title == "" || len(title) > maxTitleLength {
		return
	}

	// Description must be between 1 maxDescriptionLength long
	if len(description) > maxDescriptionLength {
		return
	}

	id := exhibition.itemCounter.Next()
	i := &Item{
		id:          id,
		title:       title,
		description: description,
		pkgpath:     pkgpath,
		blockNum:    runtime.ChainHeight(),
		upvote:      &addrset.Set{},
		downvote:    &addrset.Set{},
	}

	exhibition.items.Set(pkgpath, i)
	exhibition.itemsSortedByCreation.Set(getCreationSortKey(i.blockNum, i.id), i)
	exhibition.itemsSortedByUpvotes.Set(getVoteSortKey(i.upvote.Size(), i.id), i)
	exhibition.itemsSortedByDownvotes.Set(getVoteSortKey(i.downvote.Size(), i.id), i)

	chain.Emit("Registration")
}

func Upvote(cur realm, pkgpath string) {
	rawItem, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem)
	}

	item := rawItem.(*Item)
	caller := runtime.PreviousRealm().Address()

	if item.upvote.Has(caller) {
		panic(ErrDoubleUpvote)
	}

	if _, exists := exhibition.itemsSortedByUpvotes.Remove(getVoteSortKey(item.upvote.Size(), item.id)); !exists {
		panic("error removing old upvote entry")
	}

	item.upvote.Add(caller)

	exhibition.itemsSortedByUpvotes.Set(getVoteSortKey(item.upvote.Size(), item.id), item)
}

func Downvote(cur realm, pkgpath string) {
	rawItem, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem)
	}

	item := rawItem.(*Item)
	caller := runtime.PreviousRealm().Address()

	if item.downvote.Has(caller) {
		panic(ErrDoubleDownvote)
	}

	if _, exist := exhibition.itemsSortedByDownvotes.Remove(getVoteSortKey(item.downvote.Size(), item.id)); !exist {
		panic("error removing old downvote entry")

	}

	item.downvote.Add(caller)

	exhibition.itemsSortedByDownvotes.Set(getVoteSortKey(item.downvote.Size(), item.id), item)
}

func Pause(cur realm) error {
	return Pausable.Pause()
}

func Unpause(cur realm) error {
	return Pausable.Unpause()
}

func Delete(cur realm, pkgpath string) {
	if !Ownable.Owned() {
		panic(ownable.ErrUnauthorized)
	}

	i, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem)
	}

	item := i.(*Item)
	upvoteKey := getVoteSortKey(item.upvote.Size(), item.id)
	downvoteKey := getVoteSortKey(item.downvote.Size(), item.id)

	if _, removed := exhibition.items.Remove(pkgpath); !removed {
		panic(ErrNoSuchItem)
	}

	if _, removed := exhibition.itemsSortedByUpvotes.Remove(upvoteKey); !removed {
		panic(ErrNoSuchItem)
	}

	if _, removed := exhibition.itemsSortedByDownvotes.Remove(downvoteKey); !removed {
		panic(ErrNoSuchItem)
	}

	if _, removed := exhibition.itemsSortedByCreation.Remove(getCreationSortKey(item.blockNum, item.id)); !removed {
		panic(ErrNoSuchItem)
	}
}

func getVoteSortKey(votes int, id seqid.ID) string {
	votesStr := strconv.Itoa(votes)
	paddedVotes := strings.Repeat("0", 10-len(votesStr)) + votesStr
	return paddedVotes + ":" + strconv.FormatUint(uint64(id), 10)
}

func getCreationSortKey(blockNum int64, id seqid.ID) string {
	blockNumStr := strconv.Itoa(int(blockNum))
	paddedBlockNum := strings.Repeat("0", 10-len(blockNumStr)) + blockNumStr
	return paddedBlockNum + ":" + strconv.FormatUint(uint64(id), 10)
}
