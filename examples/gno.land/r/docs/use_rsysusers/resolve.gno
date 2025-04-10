package use_rsysusers

import (
	"std"

	"gno.land/r/sys/users" // Import the user registry
)

func Render(path string) string {
	out := "# Username checker\n\n"

	// Default render
	if path == "" {
		out += `This is short example on how developers can use the gno.land user registry to resolve
usernames and addresses in their realms. Check out the example below.

---

`
		out += "Add `:{name OR address}` to the search bar to check for a name or address.\n\n"
		out += "Here are some examples:\n\n"
		out += "- [@test1](/r/docs/users:test1)\n"
		out += "- [g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5](/r/docs/users:g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5)\n"

		return out
	}

	// If the user inputted an address
	if std.Address(path).IsValid() {
		// Try resolving an address
		data := users.ResolveAddress(std.Address(path))
		if data != nil {
			out += "## Found the user you're looking for: "
			// RenderLink will return a clickable gnoweb link leading to the user's page
			out += data.RenderLink("")
			return out
		}
	}

	// Else, try resolving a name to get user data
	data, _ := users.ResolveName(path)
	if data != nil {
		out += "## Found the user you're looking for: "
		out += data.RenderLink("")
		return out
	}

	out += "## Didn't find that user :/"

	return out
}
