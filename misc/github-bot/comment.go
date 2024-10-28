package main

import "bot/client"

type Auto struct {
	Met         bool
	Description string
}
type Manual struct {
	CheckedBy   string
	Description string
}

type Comment struct {
	Auto   []Auto
	Manual []Manual
}

func onCommentUpdated(gh *client.Github) {
}
