// Package hof is the gno.land hall of fame realm.
// It contains a permanent exhibition of items in the hall of fame,
// a possibility to host temporary ones created by its admins.
package hof

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"time"
)

var (
	idCounter seqid.ID
	permanent *Exhibition
	temporary *avl.Tree
)

type (
	Exhibition struct {
		id          seqid.ID
		title       string
		description string
		startTime   time.Time // given in RFC3339
		endTime     time.Time // end time of the event, given in RFC3339
		items       *avl.Tree // submission id >
	}

	pair struct {
		deployer          std.Address
		renderDescription func() string
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

	temporary = avl.NewTree()
}

func Register(rDesc func() string) {
	submission := std.PrevRealm()
	pkgpath := submission.PkgPath()

	// must not exist and must be called from code
	if permanent.items.Has(pkgpath) || submission.IsUser() {
		return
	}

	permanent.items.Set(pkgpath, pair{
		deployer:          std.GetOrigCaller(),
		renderDescription: rDesc,
	})

}

func Render(_ string) string {
	out := ""

	permanent.items.Iterate("", "", func(key string, value interface{}) bool {
		rFunc := value.(rndr)
		out += rFunc("")

		return false
	})

	return out
}
