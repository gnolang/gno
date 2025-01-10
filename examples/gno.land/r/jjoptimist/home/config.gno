package home

import (
    "std"
    "gno.land/p/demo/ownable"
)

var config = struct {
    title       string
    description string
    github      string
}{
    title:       "JJOptimist's Home Realm 🏠",
    description: "Exploring Gno and building on-chain",
    github:      "jjoptimist",
}

var o = ownable.NewWithAddress(std.Address("g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj"))

func GetConfig() struct { title string; description string; github string } {
    return config
}

func UpdateConfig(newTitle, newDescription, newGithub string) {
    if !o.CallerIsOwner() {
        panic("not authorized")
    }
    config.title = newTitle
    config.description = newDescription
    config.github = newGithub
}

func GetOwner() std.Address {
    return o.Owner()
} 