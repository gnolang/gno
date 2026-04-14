package home

import "gno.land/p/nt/ownable/v0"

type Config struct {
	Title       string
	Description string
	Github      string
}

var config = Config{
	Title:       "JJOptimist's Home Realm 🏠",
	Description: "Exploring Gno and building on-chain",
	Github:      "jjoptimist",
}

var Ownable = ownable.NewWithAddressByPrevious(address("g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj"))

func GetConfig() Config {
	return config
}

func UpdateConfig(_ realm, newTitle, newDescription, newGithub string) {
	Ownable.AssertOwned()
	config.Title = newTitle
	config.Description = newDescription
	config.Github = newGithub
}
