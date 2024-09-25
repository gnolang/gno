// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"time"
)

var (
	idCounter seqid.ID
	permanent *Exhibition
)

type (
	Exhibition struct {
		id          seqid.ID
		title       string
		description string
		startTime   time.Time // given in RFC3339
		endTime     time.Time // end time of the event, given in RFC3339
		items       *avl.Tree // Item id > Item
	}

	Item struct {
		pkgpath  string
		deployer std.Address
		blockNum int64
		likes    int
		dislikes int
	}
)

func init() {
	permanent = &Exhibition{
		id:          idCounter.Next(),
		title:       "Hall Of Fame",
		description: "gno.land's hall of fame!",
		startTime:   time.Now(),
		endTime:     time.Now(),
		items:       avl.NewTree(),
	}
}

func Register() {
	submission := std.PrevRealm()
	pkgpath := submission.PkgPath()

	// must not exist and must be called from code
	if permanent.items.Has(pkgpath) || submission.IsUser() {
		panic("item exists or not code call")
	}

	permanent.items.Set(
		idCounter.Next().String(),
		&Item{
			pkgpath:  pkgpath,
			deployer: std.GetOrigCaller(),
			blockNum: std.GetHeight(),
			likes:    0,
			dislikes: 0,
		})
}

func Render(_ string) string {
	out := "# Hall of Fame\n\n"

	out += permanent.Render()

	// render temp exhibitions

	return out
}

func (e Exhibition) Render() string {
	out := ufmt.Sprintf("# %s\n\n", e.title)
	out += ufmt.Sprintf("%s\n\n", e.description)

	out += "<div class='columns-3'>\n\n"

	e.items.Iterate("", "", func(key string, value interface{}) bool {
		out += "<div>"
		//out += ufmt.Sprintf("Submission #%s", key)
		out += value.(*Item).Render()
		out += "</div>"

		return false
	})

	out += "</div><!-- /columns-3 -->\n\n"

	return out
}

func (i Item) Render() string {
	out := ufmt.Sprintf("`%s`\n\n", i.pkgpath)
	out += ufmt.Sprintf("Published at block #%d\n\n", i.blockNum)
	out += ufmt.Sprintf("Publisher: %s\n\n", i.deployer.String())
	out += ufmt.Sprintf("%d 👍 ---- %d 👎\n\n", i.likes, i.dislikes)
	return out
}
