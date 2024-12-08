package users

import (
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	pusers "gno.land/p/demo/users"
	sys "gno.land/r/sys/users"
)

func Render(path string) string {
	var out string

	// Render home page
	//if path == "" {
	p := pager.NewPager(sys.GetUserStore(), 5)
	page := p.MustGetPageByPath(path)

	out += "# gno.land user registry\n"

	for _, item := range page.Items {
		user := item.Value.(*pusers.User)

		out += ufmt.Sprintf("- User #%d: [**%s**](/r/gnoland/users:%s)\n", user.Number, user.Name, user.Name)
	}

	out += "\n"
	out += page.Selector()
	out += "\n\n"
	out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

	return out

	//} XXX: pager + query params doesnt work right now
	//if !strings.HasPrefix(path, "?") {
	//	// Render specific user page
	//	user := sys.ResolveUsername(path)
	//	if user == nil {
	//		out += "Invalid username provided."
	//		return out
	//	}
	//
	//	out += user.Render()
	//}

	//return out
}
