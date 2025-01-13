package v1

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl/pager"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"github.com/gnolang/gno/examples/gno.land/p/moul/realmpath"
	"github.com/gnolang/gno/examples/gno.land/r/demo/profile"
	"github.com/gnolang/gno/examples/gno.land/r/sys/users"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
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

	// Render specific user page
	var lookup *users.UserData
	if std.Address(req.Path).IsValid() {
		lookup = users.ResolveAddress(req.Path)
	} else {
		lookup = users.ResolveName(req.Path)
	}

	if lookup == nil || lookup.IsDeleted() {
		out += "This user has been deleted or does not exist.\n"
	} else {
		out += renderUserPage(lookup)
	}

	return out
}

func renderUserPage(lookup *users.UserData) string {
	out := ufmt.Sprintf("# User %s\n", lookup.Name)
	out += profile.GetStringField(lookup.Addr, "Bio", "No bio defined.")

	return out
}
