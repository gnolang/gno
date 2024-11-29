// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"strconv"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

const (
	defaultPageSize = 20
	pathArgument    = "number"
	subPath         = "page/{" + pathArgument + "}"
)

func Render(path string) string {
	return router.Render(path)
}

func renderHelper(res *mux.ResponseWriter, req *mux.Request) {
	totalSignups := len(signups)
	res.Write("# Welcome to UserBook!\n\n")

	// Get URL parameter
	page, err := strconv.Atoi(req.GetVar("number"))
	if err != nil {
		page = 1 // render first page on bad input
	}

	// Fetch paginated signups
	fetchedSignups, endIndex := GetSignupsInRange(page, defaultPageSize)
	// Handle empty page case
	if len(fetchedSignups) == 0 {
		res.Write("No users on this page!\n\n")
		res.Write("---\n\n")
		res.Write("[Back to Page #1](/r/demo/userbook:page/1)\n\n")
		return
	}

	// Write page title
	res.Write(ufmt.Sprintf("## UserBook - Page #%d:\n\n", page))

	// Write signups
	pageStartIndex := defaultPageSize * (page - 1)
	for i, signup := range fetchedSignups {
		out := ufmt.Sprintf("#### User #%d - %s - signed up at Block #%d\n", pageStartIndex+i, signup.account, signup.height)
		res.Write(out)
	}

	res.Write("---\n\n")

	// Write UserBook info
	latestSignupIndex := totalSignups - 1
	res.Write(ufmt.Sprintf("#### Total users: %d\n", totalSignups))
	res.Write(ufmt.Sprintf("#### Latest signup: User #%d at Block #%d\n", latestSignupIndex, signups[latestSignupIndex].height))

	res.Write("---\n\n")

	// Write page number
	res.Write(ufmt.Sprintf("You're viewing page #%d", page))

	// Write navigation buttons
	var prevPage string
	var nextPage string
	// If we are on any page that is not the first page
	if page > 1 {
		prevPage = ufmt.Sprintf(" - [Previous page](/r/demo/userbook:page/%d)", page-1)
	}

	// If there are more pages after the current one
	if endIndex < totalSignups {
		nextPage = ufmt.Sprintf(" - [Next page](/r/demo/userbook:page/%d)\n\n", page+1)
	}

	res.Write(prevPage)
	res.Write(nextPage)
}
