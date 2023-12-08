package main

import (
	"std"
	"time"

	"gno.land/r/demo/tamagotchi"
)

func main() {
	tamagotchi.Reset("tamagnotchi")
	println(tamagotchi.Render(""))
}

// Output:
// # tamagnotchi 😃
//
// * age: 0
// * hunger: 50
// * happiness: 50
// * health: 50
// * sleepy: 0
//
// Actions:
// * [Feed](/r/demo/tamagotchi?help&__func=Feed)
// * [Play](/r/demo/tamagotchi?help&__func=Play)
// * [Heal](/r/demo/tamagotchi?help&__func=Heal)
// * [Reset](/r/demo/tamagotchi?help&__func=Reset)
