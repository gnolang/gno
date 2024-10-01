package hof

import (
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

const (
	MaxWidgetSize = 5
)

func Render(_ string) string {
	out := "# Hall of Fame\n\n"

	exhibitions.Iterate("", "", func(key string, value interface{}) bool {
		e := value.(*Exhibition)
		out += e.Render()
		return false
	})

	return out
}

func RenderExhibWidget(exhibID seqid.ID, itemsToRender int) (string, error) {
	if itemsToRender > MaxWidgetSize {
		return "", ErrMaxWidgetSize
	}

	if itemsToRender < 1 {
		return "", ErrMinWidgetSize
	}

	out := ""
	i := 0

	e, ok := exhibitions.Get(exhibID.String())
	if !ok {
		return "", ErrNoSuchExhibition
	}

	e.(*Exhibition).items.Iterate("", "", func(key string, value interface{}) bool {
		item := value.(*Item)

		out += ufmt.Sprintf("- %s\n", fqname.RenderLink(item.pkgpath, ""))

		i++
		return i >= itemsToRender
	})

	return out, nil

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
	out := ufmt.Sprintf("\n```\n%s\n```\n\n", i.pkgpath)
	out += ufmt.Sprintf("by %s\n\n", i.deployer.String())
	out += ufmt.Sprintf("Published at Block #%d\n\n", i.blockNum)
	out += ufmt.Sprintf(`
#### [%d 👍](/r/gnoland/hof?help&__func=Upvote&exhibID=0&itemID=%s) - [%d 👎](/r/gnoland/hof?help&__func=Downvote&exhibID=0&itemID=%s)`, i.upvote.Size(), i.id.String(), i.downvote.Size(), i.id.String())

	return out
}
