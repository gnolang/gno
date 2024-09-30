package hof

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
)

func Render(_ string) string {
	out := "# Hall of Fame\n\n"

	out += permanent.Render()

	// todo render temp exhibitions

	return out
}

func (e Exhibition) Render() string {
	out := ufmt.Sprintf("## %s\n\n", e.title)
	out += ufmt.Sprintf("%s\n\n", e.description)

	out += "<div class='columns-2'>\n\n"

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
#### [%d 👍](/r/gnoland/hof?help&__func=Upvote&exhibID=0&itemID=%s) - [%d 👎](/r/gnoland/hof?help&__func=Downvote&exhibID=0&itemID=%s)`, i.upvote.Size(), i.id.String(), i.downvote.Size(), i.id.String())

	return out
}
