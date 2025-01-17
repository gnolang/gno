package v1

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"

	"gno.land/r/demo/profile"
	"gno.land/r/sys/users"
)

var (
	connectUrlBase       = "https://gno.studio/connect/view/gno.land/r/gnoland/users/v1?network=" + std.GetChainID()
	connectRegisterUrl   = connectUrlBase + "#Register"
	connectUpdateNameUrl = connectUrlBase + "#UpdateName"
)

func Render(path string) string {
	req := realmpath.Parse(path)

	if req.Path == "" {
		return renderHomePage(path)
	}

	// Otherwise, render the user page
	return renderUserPage(req.Path)
}

func renderHomePage(path string) string {
	var out string

	out += "# gno.land user registry\n"

	out += renderIntroParagraph()

	out += md.H2("User list")

	p := pager.NewPager(users.GetReadOnlyNameStore(), 20, false)
	page := p.MustGetPageByPath(path)
	for _, item := range page.Items {
		data := item.Value.(*users.UserData)

		// Skip previous names
		if item.Key != data.Name() {
			continue
		}

		var displayKey string
		if data.IsDeleted() {
			displayKey = md.Strikethrough(item.Key)
		} else {
			displayKey = ufmt.Sprintf("**%s**", item.Key)
		}

		out += ufmt.Sprintf("- User [%s](/r/gnoland/users/v1:%s)\n", displayKey, item.Key)
	}

	//p := pager.NewPager(users.GetReadonlyAddrStore(), 20, false)
	//page := p.MustGetPageByPath(path)
	//for _, item := range page.Items {
	//	data := item.Value.(*users.UserData)
	//	if data.IsDeleted() {
	//		out += ufmt.Sprintf("- User [%s](/r/gnoland/users/v1:%s)\n", md.Strikethrough(item.Key), item.Key)
	//	} else {
	//		out += ufmt.Sprintf("- User [**%s**](/r/gnoland/users/v1:%s)\n", item.Key, item.Key)
	//	}
	//}

	out += "\n"
	out += page.Picker()
	out += "\n\n"
	out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

	return out
}

func renderIntroParagraph() string {
	out := md.Paragraph("Welcome to the gno.land user registry (v1). Please register a username.")
	out += md.Paragraph(`Registering a username grants the registering address the right to deploy packages and realms
under that username’s namespace. For example, if an address registers the username ` + md.InlineCode("gnome123") + `, it 
will gain permission to deploy packages and realms to package paths with the pattern ` + md.InlineCode("gno.land/{p,r}/gnome123/*") + `.`)

	out += md.Paragraph("In V1, usernames must follow these rules, in order to prevent username squatting:")
	items := []string{
		"Must start with 3 characters",
		"Must end with 3 numbers",
		"Have a maximum length of 20 characters",
		"With the only special character allowed being `_`",
	}
	out += md.BulletList(items)

	out += "\n\n"
	out += md.Paragraph("In later versions of the registry, vanity usernames will be allowed through specific mechanisms.")

	out += md.H2("Actions\n\n")

	items = []string{
		"Register: " + ufmt.Sprintf("[gnokey](%s) - [Connect](%s)\n\n", txlink.Call("Register"), connectRegisterUrl),
		"Update Name: " + ufmt.Sprintf("[gnokey](%s) - [Connect](%s)\n\n", txlink.Call("UpdateName"), connectUpdateNameUrl),
	}

	out += md.BulletList(items)

	out += md.HorizontalRule()
	out += "\n\n"

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
			"Note: You searched for `%s`, which is a previous name of [`%s`](/r/gnoland/users/v1:%s).",
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
