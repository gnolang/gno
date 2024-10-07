package hof

import (
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

const (
	likesBar = "#### [%d 👍](/r/demo/hof?help&__func=Upvote&pkgpath=%s) - [%d 👎](/r/demo/hof?help&__func=Downvote&pkgPath=%s)\n\n"
	delBar   = "[Delete](/r/demo/hof?help&__func=Delete&pkgpath=%s)\n\n"
)

func Render(path string) string {
	out := "# Hall of Fame\n\n"
	out += exhibition.Render(path == "dashboard")

	return out
}

func (e Exhibition) Render(dashboard bool) string {
	out := ufmt.Sprintf("%s\n\n", e.description)

	if e.items.Size() == 0 {
		out += "No items in this exhibition currently.\n\n"
		return out
	}

	out += "<div class='columns-2'>\n\n"

	e.itemsSorted.ReverseIterate("", "", func(key string, value interface{}) bool {
		out += "<div>\n\n"
		id, _ := seqid.FromString(key)
		out += ufmt.Sprintf("### Submission #%d\n\n", int(id))
		out += value.(*Item).Render(dashboard)
		out += "</div>"

		return false
	})

	out += "</div><!-- /columns-2 -->\n\n"

	return out
}

func (i Item) Render(dashboard bool) string {
	out := ufmt.Sprintf("\n```\n%s\n```\n\n", i.pkgpath)
	out += ufmt.Sprintf("by %s\n\n", i.deployer.String())
	out += ufmt.Sprintf("[View realm](%s)\n\n", fqname.ShortPath(i.pkgpath))
	out += ufmt.Sprintf("Published at Block #%d\n\n", i.blockNum)
	out += ufmt.Sprintf(
		likesBar,
		i.upvote.Size(),
		i.pkgpath,
		i.downvote.Size(),
		i.pkgpath,
	)

	if dashboard {
		out += ufmt.Sprintf(
			delBar,
			i.pkgpath,
		)
	}

	return out
}

func RenderExhibWidget(itemsToRender int) string {
	if itemsToRender < 1 {
		return ""
	}

	if exhibition.items.Size() == 0 {
		return ""
	}

	out := ""
	i := 0
	exhibition.items.Iterate("", "", func(key string, value interface{}) bool {
		item := value.(*Item)

		out += ufmt.Sprintf("- %s\n", fqname.RenderLink(item.pkgpath, ""))

		i++
		return i >= itemsToRender
	})

	return out
}
