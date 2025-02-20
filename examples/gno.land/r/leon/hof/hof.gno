// Package hof is the hall of fame realm.
// The Hall of Fame is an exhibition that holds items. Users can add their realms to the Hall of Fame by
// importing the Hall of Fame realm and calling hof.Register() from their init function.
package hof

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/addrset"
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
		items                  *avl.Tree // pkgPath > &Item
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

	Ownable = ownable.NewWithAddress(config.OwnableMain.Owner()) // OrigSendOwnable?
	Pausable = pausable.NewFromOwnable(Ownable)
}

// Register registers your realm to the Hall of Fame
// Should be called from within code
func Register(title, description string) {
	if Pausable.IsPaused() {
		return
	}

	submission := std.PreviousRealm()
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
	if len(title) == 0 || len(title) > maxTitleLength {
		return
	}

	// Description must be between 1 maxDescriptionLength long
	if len(description) == 0 || len(description) > maxDescriptionLength {
		return
	}

	id := exhibition.itemCounter.Next()
	i := &Item{
		id:          id,
		title:       title,
		description: description,
		pkgpath:     pkgpath,
		blockNum:    std.ChainHeight(),
		upvote:      &addrset.Set{},
		downvote:    &addrset.Set{},
	}

	exhibition.items.Set(pkgpath, i)
	exhibition.itemsSortedByCreation.Set(generateCreationSortKey(i.blockNum, i.id), i)
	exhibition.itemsSortedByUpvotes.Set(generateVoteSortKey(i.upvote.Size(), i.id), i)
	exhibition.itemsSortedByDownvotes.Set(generateVoteSortKey(i.downvote.Size(), i.id), i)

	std.Emit("Registration")
}

func Upvote(pkgpath string) {
	rawItem, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem)
	}

	item := rawItem.(*Item)
	caller := std.PreviousRealm().Address()

	if item.upvote.Has(caller) {
		panic(ErrDoubleUpvote)
	}

	exhibition.itemsSortedByUpvotes.Remove(generateVoteSortKey(item.upvote.Size(), item.id))
	if !item.upvote.Add(caller) {
		panic(ErrDoubleUpvote)
	}
	exhibition.itemsSortedByUpvotes.Set(generateVoteSortKey(item.upvote.Size(), item.id), item)
}

func Downvote(pkgpath string) {
	rawItem, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem)
	}

	item := rawItem.(*Item)
	caller := std.PreviousRealm().Address()

	if item.downvote.Has(caller) {
		panic(ErrDoubleDownvote)
	}

	exhibition.itemsSortedByDownvotes.Remove(generateVoteSortKey(item.downvote.Size(), item.id))
	if !item.downvote.Add(caller) {
		panic(ErrDoubleDownvote)
	}
	exhibition.itemsSortedByDownvotes.Set(generateVoteSortKey(item.downvote.Size(), item.id), item)
}

func Delete(pkgpath string) {
	if !Ownable.CallerIsOwner() {
		panic(ownable.ErrUnauthorized)
	}

	i, ok := exhibition.items.Get(pkgpath)
	if !ok {
		panic(ErrNoSuchItem)
	}

	item := i.(*Item)
	upvoteKey := generateVoteSortKey(item.upvote.Size(), item.id)
	downvoteKey := generateVoteSortKey(item.downvote.Size(), item.id)

	if _, removed := exhibition.items.Remove(pkgpath); !removed {
		panic(ErrNoSuchItem)
	}

	if _, removed := exhibition.itemsSortedByUpvotes.Remove(upvoteKey); !removed {
		panic(ErrNoSuchItem)
	}

	if _, removed := exhibition.itemsSortedByDownvotes.Remove(downvoteKey); !removed {
		panic(ErrNoSuchItem)
	}

	if _, removed := exhibition.itemsSortedByCreation.Remove(generateCreationSortKey(item.blockNum, item.id)); !removed {
		panic(ErrNoSuchItem)
	}
}

func generateVoteSortKey(votes int, id seqid.ID) string {
	votesStr := strconv.Itoa(votes)
	paddedVotes := strings.Repeat("0", 10-len(votesStr)) + votesStr
	return paddedVotes + ":" + strconv.FormatUint(uint64(id), 10)
}

func generateCreationSortKey(blockNum int64, id seqid.ID) string {
	blockNumStr := strconv.FormatInt(blockNum, 10)
	paddedBlockNum := strings.Repeat("0", 10-len(blockNumStr)) + blockNumStr
	return paddedBlockNum + ":" + strconv.FormatUint(uint64(id), 10)
}

// func generateUpvoteSortKey(item *Item) string {
// 	upvotes := item.upvote.Size()
// 	return ufmt.Sprintf("%010d:%s", upvotes, item.id.String())
// }
