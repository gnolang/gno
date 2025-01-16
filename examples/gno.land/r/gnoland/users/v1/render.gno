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

	if req.Path == "" {
		out += renderHomePage(req.Path)
	}

	// Render single user page
	data, isLatest, isName := resolveUser(req.Path)
	if data == nil {
		return render404()
	}

	// Otherwise, render the user page
	return renderUserPage(data, req.Path, isName, isLatest)
}

func renderHomePage(path string) string {
	var out string
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

// resolveUser resolves the user based on the path, determining if it's a name or address
func resolveUser(path string) (*users.UserData, bool, bool) {
	if std.Address(path).IsValid() {
		return users.ResolveAddress(std.Address(path)), false, false
	}

	data, isLatest := users.ResolveName(path)
	return data, isLatest, true
}

// renderUserPage generates the user page based on user data and path
func renderUserPage(data *users.UserData, query string, isName, isLatest bool) string {
	var out string

	out += md.H1("User " + data.Name())
	out += ufmt.Sprintf("Address: %s\n\n", data.Addr().String())

	if isName && !isLatest {
		out += md.Paragraph(ufmt.Sprintf(
			"Note: `%s` is a previous name of `%s`.",
			query, data.Name()))
	}

	out += md.H2("Bio")
	out += profile.GetStringField(data.Addr(), "Bio", "No bio defined.")
	out += "\n\n"

	return out
}

// render404 generates a 404 error page
func render404() (out string) {
	out += md.H1("404")
	out += "This user does not exist or has been deleted.\n"
	return out
}
