package home

import (
	"chain/runtime"

	"gno.land/r/sacha/config"
)

func UpdateTitle(cur realm, newTitle string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	title = newTitle
}

func UpdateBanner(cur realm, newBanner string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	banner = newBanner
}

func UpdateStory(cur realm, newStory string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	story = newStory
}

func UpdateArt(cur realm, newArt string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	art = newArt
}

func UpdateFooter(cur realm, newGitHubUsername string, newPocInnovationName string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	githubUsername = newGitHubUsername
	pocInnovationName = newPocInnovationName
}

func RemoveGame(cur realm, name string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := game[name]; exists {
		delete(game, name)
	} else {
		panic("Game not found")
	}
}

func AddGame(cur realm, name string, title string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	game[name] = title
}
