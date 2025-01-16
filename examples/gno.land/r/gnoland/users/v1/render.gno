package v1

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/realmpath"

	"gno.land/r/demo/profile"
	"gno.land/r/sys/users"
)

func Render(path string) string {
	var out string
	req := realmpath.Parse(path)

	// Render home page
	if req.Path == "" {
		p := pager.NewPager(users.GetReadonlyNameStore(), 50, false)
		page := p.MustGetPageByPath(path)

		out += "# gno.land user registry\n"

		for _, item := range page.Items {
			if !item.Value.(*users.UserData).IsDeleted() {
				out += ufmt.Sprintf("- User [**%s**](/r/gnoland/users/v1:%s)\n", item.Key, item.Key)
			}
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
		lookup = users.ResolveAddress(std.Address(req.Path))
	} else {
		lookup, _ = users.ResolveName(req.Path)
	}

	if lookup == nil {
		out += "This user has been deleted or does not exist.\n"
	} else {
		out += renderUserPage(lookup)
	}

	return out
}

func renderUserPage(lookup *users.UserData) string {
	out := ufmt.Sprintf("# User %s\n", lookup.Name())
	out += profile.GetStringField(lookup.Addr(), "Bio", "No bio defined.")

	return out
}
