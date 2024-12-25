package home

import (
    "std"
)

var (
    
    OWNER_ADDR = std.Address("g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj")

    owner std.Address

    config = struct {
        title       string
        description string
        github     string
    }{
        title:       "JJOptimist's Home Realm 🏠",
        description: "Exploring Gno and building on-chain",
        github:      "jjoptimist",
    }
)

func init() {
    owner = OWNER_ADDR
}

func IsOwner(addr std.Address) bool {
    return addr == owner
}

func GetConfig() struct {
    title       string
    description string
    github     string
} {
    return config
}

func UpdateConfig(newTitle, newDescription, newGithub string) {
    if !IsOwner(std.GetOrigCaller()) {
        panic("not authorized")
    }
    config.title = newTitle
    config.description = newDescription
    config.github = newGithub
}

func GetOwner() std.Address {
    return owner
} 