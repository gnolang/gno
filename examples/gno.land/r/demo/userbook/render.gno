// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"strconv"

	"gno.land/r/sys/users"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/ufmt"
)

const usersLink = "/r/gnoland/users/v1"

func Render(path string) string {
	p := pager.NewPager(signupsTree, 20, true)
	page := p.MustGetPageByPath(path)

	out := "# Welcome to UserBook!\n\n"

	out += ufmt.Sprintf("## [Click here to sign up!](%s)\n\n", txlink.Call("SignUp"))
	out += "---\n\n"

	for _, item := range page.Items {
		signup := item.Value.(*Signup)
		user := signup.address_XXX.String()

		if data := users.ResolveAddress(signup.address_XXX); data != nil {
			user = ufmt.Sprintf("[%s](%s:%s)", data.Name(), usersLink, data.Name())
		}

		out += ufmt.Sprintf("- **User #%d - %s - signed up on %s**\n\n", signup.ordinal, user, signup.timestamp.Format("January 2 2006, 03:04:04 PM"))
	}

	out += "---\n\n"
	out += "**Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "**\n\n"
	out += page.Picker(path)
	return out
}
