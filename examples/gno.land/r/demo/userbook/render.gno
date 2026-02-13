// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"strconv"

	"gno.land/r/sys/users"

	"gno.land/p/demo/ui/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/ufmt"
)

const usersLink = "/r/gnoland/users/v1"

func Render(path string) string {
	p := pager.NewPager(signupsTree, 20, true)
	page := p.MustGetPageByPath(path)

	// Build the page using the md package
	doc := md.New().
		H1("Welcome to UserBook!").
		H2(md.Link("Click here to sign up!", txlink.Call("SignUp"))).
		HR()

	// Add each signup entry
	for _, item := range page.Items {
		signup := item.Value.(*Signup)
		user := signup.address_XXX.String()

		if data := users.ResolveAddress(signup.address_XXX); data != nil {
			user = md.Link(data.Name(), usersLink+":"+data.Name())
		}

		entry := ufmt.Sprintf("User #%d - %s - signed up on %s",
			signup.ordinal, user, signup.timestamp.Format("January 2 2006, 03:04:04 PM"))
		doc.P(md.Bold(entry))
	}

	doc.HR()
	doc.P(md.Bold("Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages)))
	doc.Write(page.Picker(path))

	return doc.String()
}
