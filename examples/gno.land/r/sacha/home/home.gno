package home

import (
	"gno.land/p/moul/md"
)

func Render(_ string) string {
	out := "![banner](https://media.licdn.com/dms/image/v2/D4E16AQEoUC0tW-XFew/profile-displaybackgroundimage-shrink_350_1400/profile-displaybackgroundimage-shrink_350_1400/0/1715027502009?e=1756339200&v=beta&t=4-mxY-Zzq36GQFhzSoDJX9gcjxYfchbgAxBO5Gjfmqo)\n"
	out += "# Sacha's Space 🏗️\n"
	out += "In the quiet layers of gno.land, a gnome brushed against a line of forgotten code. No one saw him. No one knows what he changed. But sometimes, in the logs...\n\n\"gnonk.\"\n\nJust once. Then silence.\n";

	out += "## "
	out += md.Link("🪙 Flip a coin!", "coinflip")
	out += "\n"

	return out
}
