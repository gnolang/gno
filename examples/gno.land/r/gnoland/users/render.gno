package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl/pager"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"github.com/gnolang/gno/examples/gno.land/p/moul/realmpath"
	"github.com/gnolang/gno/examples/gno.land/p/sys/users"
	"github.com/gnolang/gno/examples/gno.land/r/sys/users"
	"strconv"
)

func Render(path string) string {
	var out string
	req := realmpath.Parse(path)

	// Render home page
	if req.Path == "" {
		p := pager.NewPager(users.GetReadonlyStore(), 5, false)
		page := p.MustGetPageByPath(path)

		out += "# gno.land user registry\n"

		for _, item := range page.Items {
			out += ufmt.Sprintf("- User [**%s**](/r/gnoland/users:%s)\n", item.Key, item.Key)
		}

		out += "\n"
		out += page.Picker()
		out += "\n\n"
		out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

		return out
	}

	// specific user page
	lookup := users.ResolveName(req.Path)
	out += renderUserPage(lookup)

	return out
}

func renderUserPage(lookup users.UserLookup) string {
	// cheeck if alias
	// if alias display original name as well
	out := ufmt.Sprintf("# User %s\n", lookup.Name)

	return out
}
