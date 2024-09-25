// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"std"
	"time"
)

var (
	exhibCounter seqid.ID
	permanent    *Exhibition
)

type (
	Exhibition struct {
		id          seqid.ID // fixed
		itemCounter seqid.ID //
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
}

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

func Render(_ string) string {
	out := "# Hall of Fame\n\n"

	out += permanent.Render()

	// todo render temp exhibitions

	return out
}

func (e Exhibition) Render() string {
	//out := ufmt.Sprintf("## %s\n\n", e.title)
	//out += ufmt.Sprintf("%s\n\n", e.description)

	out := "<div class='columns-2'>\n\n"

	e.items.ReverseIterate("", "", func(key string, value interface{}) bool {
		out += "<div>\n\n"
		id, _ := seqid.FromString(key)
		out += ufmt.Sprintf("### Submission #%d\n\n", int(id))
		out += value.(*Item).Render()
		out += "</div>"

		return false
	})

	out += "</div><!-- /columns-2 -->\n\n"

	return out
}

func (i Item) Render() string {
	//subpath := i.pkgpath[strings.Index(i.pkgpath, "/"):]

	out := ufmt.Sprintf("\n```\n%s\n```\n\n", i.pkgpath)
	out += ufmt.Sprintf("by %s\n\n", i.deployer.String())
	out += ufmt.Sprintf("Published at Block #%d\n\n", i.blockNum)
	out += ufmt.Sprintf(`
#### [%d 👍](/r/gnoland/hof?help&__func=Upvote&exhibitionID=0&itemID=%s) - [%d 👎](/r/gnoland/hof?help&__func=Downvote&exhibitionID=0&itemID=%s)`, i.upvote.Size(), i.id.String(), i.downvote.Size(), i.id.String())

	return out
}

func Upvote(exhibitionID, itemID seqid.ID) error {
	if exhibitionID == seqid.ID(0) {
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

func Downvote(exhibitionID, itemID seqid.ID) error {
	if exhibitionID == seqid.ID(0) {
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
