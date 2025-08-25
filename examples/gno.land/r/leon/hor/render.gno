package hor

import (
	"net/url"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/fqname"
	"gno.land/p/nt/ufmt"

	"gno.land/r/sys/users"
)

const colsPerRow = 3

func Render(path string) string {
	out := md.H1("The Hall of Realms\n\n")

	if !strings.Contains(path, "hall") && !strings.Contains(path, "dashboard") {
		out += renderAbout()
		return out
	}

	dashboardEnabled := strings.Contains(path, "dashboard")
	if dashboardEnabled {
		out += renderDashboard()
		out += renderActions(path)
		return out
	}

	out += renderActions(path)
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
		str[i] += itemValue.Render(dashboard)
	}

	out += md.ColumnsN(str, 3, true)
	out += md.H3(page.Picker(path))

	return out
}

func (i Item) Render(dashboard bool) string {
	out := ufmt.Sprintf("### [%s](%s)\n",
		i.title,
		strings.TrimPrefix(i.pkgpath, "gno.land"),
	)

	if i.description == "" {
		i.description = "_No description provided._\n"
	}
	out += ufmt.Sprintf("%s\n\n", i.description)

	namespace := strings.Split(i.pkgpath, "/")[2]
	user, _ := users.ResolveAny(namespace)
	if user != nil {
		namespace = user.RenderLink("")
	}
	out += ufmt.Sprintf("by %s\n\n", namespace)

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
	out := md.H3("Dashboard\n\n")
	out += ufmt.Sprintf("Total submissions: %d\n\n", exhibition.items.Size())

	out += ufmt.Sprintf("Exhibition admin: %s\n\n", Ownable.Owner().String())

	if !Pausable.IsPaused() {
		out += md.Link("Pause exhibition", txlink.Call("Pause"))
	} else {
		out += md.Link("Unpause exhibition", txlink.Call("Unpause"))
	}

	out += "\n\n"
	return out
}

func RenderExhibWidget(itemsToRender int) string {
	if itemsToRender < 1 {
		return ""
	}

	if exhibition.items.Size() == 0 {
		return "No items in the Hall of Realms.\n\n"
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
		return e.itemsSortedByCreation
	}
}

func renderAbout() string {
	out := `
Welcome, gnomes!

The Hall of Realms is a simple & permissionless dashboard for gnomes to share
their work with the community.

Here, anyone is welcome to submit their own code. This realm utilizes a common
Gno pattern - the registry pattern - to allow anyone to programmatically submit
their work.

Simply import the Hall of Realms in your code, and call the "Register()" function
inside your realm init, as shown below:

"""go
package myrealm

import "gno.land/r/leon/hor"

func init() {
	hor.Register(cross, "My Gnome App", "This is my submission to the Hall of Realms.")
}
...
"""

## [Visit The Hall ->](/r/leon/hor:hall)

`
	out = strings.ReplaceAll(out, "\"", "`")

	return out
}

func renderActions(path string) string {
	out := md.HorizontalRule()
	out += md.Link("Reset Sort", "?") + " | "
	out += md.Link("Sort by Upvotes", "?sort=upvotes") + " | "
	out += md.Link("Sort by Downvotes", "?sort=downvotes") + " | "
	out += md.Link("Sort by Most Recent", "?sort=creation") + " | "
	out += md.Link("Sort by Oldest", "?sort=oldest") + " | "

	if !strings.Contains(path, "dashboard") {
		out += md.Link("Dashboard", "/r/leon/hor:dashboard")
	} else {
		out += md.Link("Dashboard off", "/r/leon/hor:hall")
	}

	out += " | "
	out += md.Link("About", "/r/leon/hor") + "\n\n"

	out += md.HorizontalRule()

	return out
}
