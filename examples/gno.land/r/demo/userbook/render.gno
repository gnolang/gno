// Package userbook demonstrates a small userbook system working with gnoweb
package userbook

import (
	"sort"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
)

func Render(path string) string {
	p := pager.NewPager(signupsTree, 2)
	page := p.MustGetPageByPath(path)

	out := "# Welcome to UserBook!\n\n"

	out += ufmt.Sprintf("## [Click here to sign up!](%s)\n\n", txlink.Call("SignUp"))
	out += "---\n\n"

	var sorted sortedSignups
	for _, item := range page.Items {
		sorted = append(sorted, item.Value.(*Signup))
	}

	sort.Sort(sorted)

	for _, item := range sorted {
		out += ufmt.Sprintf("- **User #%d - %s - signed up on %s**\n\n", item.ordinal, item.address.String(), item.timestamp.Format("02-01-2006 15:04:05"))
	}

	out += "---\n\n"
	out += "**Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "**\n\n"
	out += page.Selector() // Repeat selector for ease of navigation
	return out
}

type sortedSignups []*Signup

func (s sortedSignups) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedSignups) Len() int {
	return len(s)
}

func (s sortedSignups) Less(i, j int) bool {
	return s[i].timestamp.Before(s[j].timestamp)
}
