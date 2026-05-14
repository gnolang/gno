package buttons

import (
	"chain/runtime"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt/v0"
)

var (
	motd       = "The Initial Message\n\n"
	lastCaller address
)

func UpdateMOTD(_ realm, newmotd string) {
	motd = newmotd
	lastCaller = runtime.PreviousRealm().Address()
}

func Render(path string) string {
	if path == "motd" {
		out := "# Message of the Day:\n\n"
		out += "---\n\n"
		out += "# " + motd + "\n\n"
		out += "---\n\n"
		link := txlink.Call("UpdateMOTD", "newmotd", "Message!") // "/r/docs/buttons$help&func=UpdateMOTD&newmotd=Message!"
		out += ufmt.Sprintf("Click **[here](%s)** to update the Message of The Day!\n\n", link)
		out += "[Go back to home page](/r/docs/buttons)\n\n"
		out += "Last updated by " + lastCaller.String()

		return out
	}

	out := `# Buttons

Users can create simple hyperlink buttons to view specific realm pages and
do specific realm actions, such as calling a specific function with some arguments.

The foundation for this functionality are markdown links; for example, you can
click...
` + "\n## [here](/r/docs/buttons:motd)\n" + `...to view this realm's message of the day.`

	return out
}
