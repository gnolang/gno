package home

import (
    "time"
    "std"
    "strconv"
)

var (
    creation time.Time
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

func getGnomeArt(height int64) string {
    var art string
    if height%7 == 0 {
        art = gnomeArt4  // winking gnome
    } else if height%5 == 0 {
        art = gnomeArt3  // starry-eyed gnome
    } else if height%3 == 0 {
        art = gnomeArt2  // happy gnome
    } else if height%2 == 0 {
        art = gnomeArt1  // regular gnome
    } else {
        art = gnomeArt1  // default
    }
    return "```\n" + art + "\n```\n"
}

func init() {
    creation = time.Now()
}

func Render(path string) string {
    cfg := GetConfig()
    height := std.GetHeight()
    
    output := "# " + cfg.title + "\n\n"
    
    output += "## About Me\n"
    output += "- 👋 Hi, I'm JJOptimist\n"
    output += getGnomeArt(height)
    output += "- 🌱 " + cfg.description + "\n"
    
    output += "## Contact\n"
    output += "- 📫 GitHub: [" + cfg.github + "](https://github.com/" + cfg.github + ")\n"
    
    output += "\n---\n"
    output += "_Realm created: " + creation.Format("2006-01-02 15:04:05 UTC") + "_\n"
    output += "_Owner: " + GetOwner().String() + "_\n"
    output += "_Current Block Height: " + strconv.Itoa(int(height)) + "_"
    
    return output
} 