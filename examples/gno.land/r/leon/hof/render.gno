package hof

import (
	"net/url"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"
)

const (
	colsPerRow = 3
)

func Render(path string) string {
	out := md.H1("Hall of Fame\n\n")
	out += md.Link("Reset Sort", "?") + " | "
	out += md.Link("Sort by Upvotes", "?sort=upvotes") + " | "
	out += md.Link("Sort by Downvotes", "?sort=downvotes") + " | "
	out += md.Link("Sort by Most Recent", "?sort=creation") + " | "
	out += md.Link("Sort by Oldest", "?sort=oldest") + "\n\n"

	dashboardEnabled := path == "dashboard"

	if dashboardEnabled {
		out += renderDashboard()
	}

	out += exhibition.Render(path, dashboardEnabled)

	return out
}

func (e Exhibition) Render(path string, dashboard bool) string {
	tree := getTreeByPath(&e, path)

	u, _ := url.Parse(path)
	reversed := u.Query().Get("sort") != "oldest"

	page := pager.NewPager(tree, colsPerRow*3, reversed).MustGetPageByPath(path)
	out := ufmt.Sprintf("%s\n\n", e.description)

	if e.items.Size() == 0 {
		out += "No items in this exhibition currently.\n\n"
		return out
	}

	str := make([]string, len(page.Items))
	for i, item := range page.Items {
		itemValue := item.Value.(*Item)
		str[i] += ufmt.Sprintf("### %d - %s\n", int(itemValue.id), itemValue.title)
		str[i] += itemValue.Render(dashboard)
	}

	out += md.ColumnsN(str, colsPerRow, true)
	out += page.Picker(path)

	return out
}

func (i Item) Render(dashboard bool) string {
	out := ufmt.Sprintf("\n%s\n\n", md.CodeBlock(i.pkgpath))
	//out := ufmt.Sprintf("\n%s\n\n", i.pkgpath)
	if i.description == "" {
		i.description = "_No description provided._\n"
	}
	out += ufmt.Sprintf("%s\n\n", i.description)

	namespace := strings.Split(i.pkgpath, "/")[2]
	user := users.ResolveAny(namespace)
	if user != nil {
		namespace = user.RenderLink("")
	}
	out += ufmt.Sprintf("by %s\n\n", namespace)

	out += md.Link("View Realm", strings.TrimPrefix(i.pkgpath, "gno.land")) + "\n\n"

	blockMsg := "Submitted via the [Monorepo](https://github.com/gnolang/gno)"
	if i.blockNum > 0 {
		blockMsg = ufmt.Sprintf("Submitted at Block #%d", i.blockNum)
	}
	out += ufmt.Sprintf("%s\n\n", blockMsg)

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
	exhibition.items.Iterate("", "", func(key string, value any) bool {
		item := value.(*Item)

		out += ufmt.Sprintf("- %s\n", fqname.RenderLink(item.pkgpath, ""))

		i++
		return i >= itemsToRender
	})

	return out
}

func getTreeByPath(e *Exhibition, path string) *avl.Tree {
	u, _ := url.Parse(path)
	switch u.Query().Get("sort") {
	case "upvotes":
		return e.itemsSortedByUpvotes
	case "downvotes":
		return e.itemsSortedByDownvotes
	case "creation":
		return e.itemsSortedByCreation
	case "oldest":
		return e.itemsSortedByCreation
	default:
		return e.items
	}
}
