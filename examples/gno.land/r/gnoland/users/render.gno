package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/moul/realmpath"
)

func Render(path string) string {
	var out string
	req := realmpath.Parse(path)

	// Render home page
	if req.Path == "" {
		// todo render home page

		//p := pager.NewPager(sys.GetUserStore(), 5)
		//page := p.MustGetPageByPath(path)
		//
		//out += "# gno.land user registry\n"
		//
		//for _, item := range page.Items {
		//	out += ufmt.Sprintf("- User #%d: [**%s**](/r/gnoland/users:%s)\n", user.Number, user.Name, user.Name)
		//}
		//
		//out += "\n"
		//out += page.Selector()
		//out += "\n\n"
		//out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"
		//
		//return out
	}

	return out
}
