package users

import "gno.land/p/nt/ufmt/v0"

func Render(_ string) string {
	out := "# r/sys/users\n\n"

	out += "`r/sys/users` is a system realm for managing user registrations.\n\n"
	out += "Users should use [`gno.land/r/gnoland/users`](/r/gnoland/users) implementations to register their usernames.\n\n"
	out += "---\n\n"

	out += "## Stats\n\n"
	out += ufmt.Sprintf("Total unique addresses registered: **%d**\n\n", addressStore.Size())
	out += ufmt.Sprintf("Total unique names registered: **%d**\n\n", nameStore.Size())
	return out
}
