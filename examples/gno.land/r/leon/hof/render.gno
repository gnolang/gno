package hof

import (
	"strings"
	"net/url"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/demo/avl"
)

const (
	pageSize = 5
)

func Render(path string) string {
	out := md.H1("Hall of Fame\n\n")
	out += md.Link("Reset Sort", "?") + " | "
	out += md.Link("Sort by Upvotes", "?sort=upvotes") + "\n\n"

	dashboardEnabled := path == "dashboard"

	if dashboardEnabled {
		out += renderDashboard()
	}

	out += exhibition.Render(path, dashboardEnabled)

	return out
}

func (e Exhibition) Render(path string, dashboard bool) string {
	out := ""

	tree := getTreeByPath(&e, path)
	page := pager.NewPager(tree, pageSize, true).MustGetPageByPath(path)

	out += ufmt.Sprintf("%s\n\n", e.description)

	if e.items.Size() == 0 {
		out += "No items in this exhibition currently.\n\n"
		return out
	}

	out += "<div class='columns-2'>\n\n"

	for _, item := range page.Items {
		out += "<div>\n\n"
		itemValue := item.Value.(*Item)
		out += md.H3(itemValue.title + "\n\n")
		out += itemValue.Render(dashboard)
		out += "</div>"
	}

	out += "</div><!-- /columns-2 -->\n\n"

	out += page.Picker()

	return out
}

func (i Item) Render(dashboard bool) string {
	out := ufmt.Sprintf("\n%s\n\n", md.CodeBlock(i.pkgpath))
	out += ufmt.Sprintf("%s\n\n", i.description)
	out += ufmt.Sprintf("by %s\n\n", strings.Split(i.pkgpath, "/")[2])
	out += md.Link("View Realm", strings.TrimPrefix(i.pkgpath, "gno.land")) + "\n\n"
	out += ufmt.Sprintf("Submitted at Block #%d\n\n", i.blockNum)

	out += md.Bold(ufmt.Sprintf("[%d👍](%s) - [%d👎](%s)",
		i.upvote.Size(), txlink.Call("Upvote", "pkgpath", i.pkgpath),
		i.downvote.Size(), txlink.Call("Downvote", "pkgpath", i.pkgpath),
	))

	if dashboard {
		out += md.Link("Delete", txlink.Call("Delete", "pkgpath", i.pkgpath))
	}

	return out
}

func renderDashboard() string {
	out := md.HorizontalRule()
	out += md.H2("Dashboard\n\n")
	out += ufmt.Sprintf("Total submissions: %d\n\n", exhibition.items.Size())

	out += ufmt.Sprintf("Exhibition admin: %s\n\n", Ownable.Owner().String())

	if !Pausable.IsPaused() {
		out += md.Link("Pause exhibition", txlink.Call("Pause"))
	} else {
		out += md.Link("Unpause exhibition", txlink.Call("Unpause"))
	}

	out += md.HorizontalRule()

	return out
}

func RenderExhibWidget(itemsToRender int) string {
	if itemsToRender < 1 {
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

func getTreeByPath(e *Exhibition, path string) *avl.Tree {
	u, _ := url.Parse(path)
	if u.Query().Get("sort") == "upvotes" {
		return e.itemsSortedByUpvotes
	}
	return e.items
}
