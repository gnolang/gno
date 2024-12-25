package home

import (
    "time"
)

var (
    creation time.Time
)

func init() {
    creation = time.Now()
}

func Render(path string) string {
    cfg := GetConfig()
    
    output := "# " + cfg.title + "\n\n"
    
    output += "## About Me\n"
    output += "- 👋 Hi, I'm JJOptimist\n"
    output += "- 🌱 " + cfg.description + "\n"
    
    output += "## Contact\n"
    output += "- 📫 GitHub: [" + cfg.github + "](https://github.com/" + cfg.github + ")\n"
    
    output += "\n---\n"
    output += "_Realm created: " + creation.Format("2006-01-02 15:04:05 MST") + "_\n"
    output += "_Owner: " + GetOwner().String() + "_"
    
    return output
} 