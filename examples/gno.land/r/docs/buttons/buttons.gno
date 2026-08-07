package buttons

import (
	"time"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	motd       = "The Initial Message"
	lastCaller address
	lastUpdate time.Time
)

// UpdateMOTD sets the message of the day and records who set it.
//
// It is a crossing function: gnoweb's <gno-form> and the $help button both
// reach it through a MsgCall. cur.Previous() is a trustworthy caller identity.
func UpdateMOTD(cur realm, newmotd string) {
	motd = newmotd
	lastCaller = cur.Previous().Address()
	lastUpdate = time.Now()
}

func Render(path string) string {
	if path == "motd" {
		out := "# Message of the Day\n\n"
		out += "---\n\n"
		out += "# " + sanitize.InlineText(motd) + "\n\n"
		out += "---\n\n"

		out += "## Submitting the transaction\n\n"
		out += "The idiomatic way to send a transaction from a rendered page is a " +
			"`<gno-form>`. The wallet collects the input and submits a `MsgCall` to " +
			"`UpdateMOTD`:\n\n"
		out += `<gno-form exec="UpdateMOTD">
<gno-input name="newmotd" placeholder="New message" required="true" />
</gno-form>` + "\n\n"

		out += "## The simpler link-only pattern\n\n"
		out += "`txlink.Call` builds a `$help` URL with the arguments pre-filled. " +
			"It produces a plain clickable link rather than a form, useful when you " +
			"want a one-click action with fixed arguments:\n\n"
		link := txlink.Call("UpdateMOTD", "newmotd", "Message!") // "/r/docs/buttons$help&func=UpdateMOTD&newmotd=Message!"
		out += ufmt.Sprintf("Click **[here](%s)** to update the Message of The Day.\n\n", link)

		out += "[Go back to home page](/r/docs/buttons)\n\n"

		if lastUpdate.IsZero() {
			out += "_Not updated yet._"
		} else {
			out += ufmt.Sprintf("Last updated by %s at %s",
				lastCaller.String(), lastUpdate.Format("2006-01-02 15:04:05"))
		}

		return out
	}

	out := `# Buttons

Realms can render clickable controls that view a specific page or submit a
transaction. There are two primitives for the transaction case:

- **` + "`<gno-form>`" + `** — the idiomatic, wallet-mediated form. It renders
  input fields and submits a ` + "`MsgCall`" + ` to the named exported function.
  Use this whenever the caller supplies arguments.
- **` + "`txlink.Call`" + `** — builds a bare ` + "`$help`" + ` link with the
  arguments baked in. No form, just a hyperlink; handy for fixed-argument
  one-click actions.

The foundation for both is a markdown link; for example, you can click...
` + "\n## [here](/r/docs/buttons:motd)\n" + `...to view this realm's message of the day, where both primitives are demonstrated.`

	return out
}
