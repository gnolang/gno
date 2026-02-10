package home

import (
	"chain/runtime"
	"strconv"
	"time"

	"gno.land/r/leon/hor"
)

const (
	gnomeArt1 = `   /\
  /  \
 ,,,,,
(o.o)
(\_/)
-"-"-`

	gnomeArt2 = `   /\
  /  \
 ,,,,,
(^.^)
(\_/)
 -"-`

	gnomeArt3 = `   /\
  /  \
 ,,,,,
(*.*)
(\_/)
"-"-"`

	gnomeArt4 = `   /\
  /  \
 ,,,,,
(o.~)
(\_/)
 -"-`
)

var creation time.Time

func getGnomeArt(height int64) string {
	var art string
	switch {
	case height%7 == 0:
		art = gnomeArt4 // winking gnome
	case height%5 == 0:
		art = gnomeArt3 // starry-eyed gnome
	case height%3 == 0:
		art = gnomeArt2 // happy gnome
	default:
		art = gnomeArt1 // regular gnome
	}
	return "```\n" + art + "\n```\n"
}

func init() {
	creation = time.Now()
	hor.Register(cross, "JJoptimist's Home Realm", "")
}

func Render(path string) string {
	height := runtime.ChainHeight()

	output := "# " + config.Title + "\n\n"

	output += "## About Me\n"
	output += "- 👋 Hi, I'm JJOptimist\n"
	output += getGnomeArt(height)
	output += "- 🌱 " + config.Description + "\n"

	output += "## Contact\n"
	output += "- 📫 GitHub: [" + config.Github + "](https://github.com/" + config.Github + ")\n"

	output += "\n---\n"
	output += "_Realm created: " + creation.Format("2006-01-02 15:04:05 UTC") + "_\n"
	output += "_Owner: " + Ownable.Owner().String() + "_\n"
	output += "_Current Block Height: " + strconv.Itoa(int(height)) + "_"

	return output
}
