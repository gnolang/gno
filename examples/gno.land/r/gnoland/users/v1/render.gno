package users

import (
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt"

	"gno.land/r/demo/profile"
	susers "gno.land/r/sys/users"
)

func Render(path string) string {
	req := realmpath.Parse(path)

	if req.Path == "" {
		return renderHomePage()
	}

	// Otherwise, render the user page
	return renderUserPage(req.Path)
}

func renderHomePage() string {
	var out string

	out += "# Gno.land User Registry\n"

	if paused {
		out += md.HorizontalRule()
		out += md.H2("This realm is paused.")
		out += md.Paragraph("Check out [`gno.land/r/gnoland/users`](/r/gnoland/users) for newer versions of the registry.")
		out += md.HorizontalRule()
	}

	out += renderIntroParagraph()

	out += md.H2("Latest registrations")
	out += RenderLatestUsersWidget(-1)

	return out
}

func renderIntroParagraph() string {
	out := md.Paragraph("Welcome to the Gno.land User Registry (v1). Please register a username.")
	out += md.Paragraph(`Registering a username grants the registering address the right to deploy packages and realms
under that username’s namespace. For example, if an address registers the username ` + md.InlineCode("gnome123") + `, it 
will gain permission to deploy packages and realms to package paths with the pattern ` + md.InlineCode("gno.land/{p,r}/gnome123/*") + `.`)

	out += md.Paragraph("In V1, usernames must follow these rules, in order to prevent username squatting:")
	items := []string{
		"Must start with 3 letters",
		"Letters must be lowercase",
		"Must end with 3 numbers",
		"Have a maximum length of 20 characters",
		"With the only special character allowed being `_`",
	}
	out += md.BulletList(items)

	out += "\n\n"
	out += md.Paragraph("In later versions of the registry, vanity usernames will be allowed through specific mechanisms.")

	if !paused {
		amount := ufmt.Sprintf("%dugnot", registerPrice)
		link := txlink.NewLink("Register")
		if registerPrice > 0 {
			link = link.SetSend(amount)
		}

		out += md.H3(ufmt.Sprintf(" [[Click here to register]](%s)", link.URL()))
		// XXX: Display registration price adjusting for dynamic GNOT price when it becomes possible.
		out += ufmt.Sprintf("Registration price: %f GNOT (%s)\n\n", float64(registerPrice)/1_000_000, amount)
	}

	out += md.HorizontalRule()
	out += "\n\n"

	return out
}

// resolveUser resolves the user based on the path, determining if it's a name or address
func resolveUser(path string) (*susers.UserData, bool, bool) {
	if address(path).IsValid() {
		return susers.ResolveAddress(address(path)), false, false
	}

	data, isLatest := susers.ResolveName(path)
	return data, isLatest, true
}

// renderUserPage generates the user page based on user data and path
func renderUserPage(path string) string {
	var out string

	// Render single user page
	data, isLatest, isName := resolveUser(path)
	if data == nil {
		out += md.H1("User not found.")
		out += "This user does not exist or has been deleted.\n"
		return out
	}

	out += md.H1("User - " + md.InlineCode(data.Name()))

	if isName && !isLatest {
		out += md.Paragraph(ufmt.Sprintf(
			"Note: You searched for `%s`, which is a previous name of [`%s`](/u/%s).",
			path, data.Name(), data.Name()))
	} else {
		out += ufmt.Sprintf("Address: %s\n\n", data.Addr().String())

		out += md.H2("Bio")
		out += profile.GetStringField(data.Addr(), "Bio", "No bio defined.")
		out += "\n\n"
		out += ufmt.Sprintf("[Update bio](%s)", txlink.Realm("gno.land/r/demo/profile").Call("SetStringField", "field", "Bio"))
		out += "\n\n"
	}

	return out
}

// RenderLatestUsersWidget renders the latest num registered users
// For num = -1, maximum number (10) will be displayed
func RenderLatestUsersWidget(num int) string {
	size := latestUsers.Size()
	if size == 0 {
		return "No registered users."
	}

	if num > size || num < 0 {
		num = size
	}

	entries := latestUsers.Entries()
	var out string

	for i := size - 1; i >= size-num; i-- {
		user := entries[i].(string)
		out += md.BulletItem(md.UserLink(user))
	}

	return out
}
