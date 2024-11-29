package buttons

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
)

var (
	number = 1
	motd   = "MOTD: The Initial Message\n\n"
)

func UpdateMOTD() {
	caller := std.PrevRealm().Addr().String()
	motd = ufmt.Sprintf("MOTD: User %s just updated the MOTD, and they were the #%d user to do it!\n\n", caller, number)
	number++
}

func Render(path string) string {
	if path == "motd" {
		out := motd

		link := txlink.URL("UpdateMOTD") //  "/r/docs/buttons$help?func=UpdateMOTD"
		out += ufmt.Sprintf("Click [this link](%s) to update the Message of The Day!\n\n", link)
		out += "[Go back to home page](/r/docs/buttons)"

		return out
	}

	out := `# Buttons

Users can create simple hyperlink buttons to view specific realm pages and 
do specific realm actions, such as calling a specific function with some arguments.

The foundation for this functionality are markdown links; for example,
check out the Message Of The Day found in this realm, [here](/r/docs/buttons:motd).
`

	return out
}
