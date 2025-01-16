package v1

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"

	"gno.land/r/demo/profile"
	"gno.land/r/sys/users"
)

func Render(path string) string {
	var out string
	req := realmpath.Parse(path)

	// Render home page
	if req.Path == "" {
		p := pager.NewPager(users.GetReadonlyAddrStore(), 50, false)
		page := p.MustGetPageByPath(path)

		out += "# gno.land user registry\n"

		for _, item := range page.Items {
			data := item.Value.(*users.UserData)
			if !data.IsDeleted() {
				out += ufmt.Sprintf("- User [**%s**](/r/gnoland/users/v1:%s)\n", data.Name(), data.Name())
			}
		}

		out += "\n"
		out += page.Picker()
		out += "\n\n"
		out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

		return out
	}

	// Render specific user page
	var data *users.UserData
	if std.Address(req.Path).IsValid() {
		data = users.ResolveAddress(std.Address(req.Path))
	} else {
		data, _ = users.ResolveName(req.Path)
	}

	if data == nil {
		out += "This user does not exist or has been deleted.\n"
	} else {
		out += renderUserPage(data)
	}

	return out
}

func renderUserPage(lookup *users.UserData) string {
	out := md.H1("User " + lookup.Name())
	out += ufmt.Sprintf("Address: %s\n\n", lookup.Addr().String())
	out += md.H2("Bio")
	out += profile.GetStringField(lookup.Addr(), "Bio", "No bio defined.")
	out += "\n\n"

	return out
}
