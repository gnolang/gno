package hof

import (
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

const (
	likesBar = "#### [%d 👍](/r/gnoland/hof?help&__func=Upvote&exhibID=%s&itemID=%s) - [%d 👎](/r/gnoland/hof?help&__func=Downvote&exhibID=%s&itemID=%s)\n\n"
	delBar   = "[Delete](/r/gnoland/hof?help&__func=Delete&exhibID=%s&itemID=%s)\n\n"
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

	e.items.ReverseIterate("", "", func(key string, value interface{}) bool {
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
	out += ufmt.Sprintf("Published at Block #%d\n\n", i.blockNum)
	out += ufmt.Sprintf(
		likesBar,
		i.upvote.Size(),
		i.exhibID.String(),
		i.id.String(),
		i.downvote.Size(),
		i.exhibID.String(),
		i.id.String(),
	)

	if dashboard {
		out += ufmt.Sprintf(
			delBar,
			i.exhibID.String(),
			i.id.String(),
		)
	}

	return out
}

func RenderExhibWidget(exhibID seqid.ID, itemsToRender int) string {

	if itemsToRender < 1 {
		return ""
	}

	e, ok := exhibitions.Get(exhibID.String())
	if !ok {
		return "", ErrNoSuchExhibition
	}

	exhib := e.(*Exhibition)
	if exhib.items.Size() == 0 {
		return "No items in this exhibition.", nil
	}

	out := ""
	i := 0

	exhib.items.Iterate("", "", func(key string, value interface{}) bool {
		item := value.(*Item)

		out += ufmt.Sprintf("- %s\n", fqname.RenderLink(item.pkgpath, ""))

		i++
		return i >= itemsToRender
	})

	return out, nil
}
