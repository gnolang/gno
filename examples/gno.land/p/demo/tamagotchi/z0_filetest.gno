package main

import (
	"os"
	"time"

	"gno.land/p/demo/tamagotchi"
)

func main() {
	t := tamagotchi.New("Gnome")

	println("\n-- INITIAL\n")
	println(t.Markdown())

	println("\n-- WAIT 20 minutes\n")
	os.Sleep(20 * time.Minute)
	println(t.Markdown())

	println("\n-- FEEDx3, PLAYx2, HEALx4\n")
	t.Feed()
	t.Feed()
	t.Feed()
	t.Play()
	t.Play()
	t.Heal()
	t.Heal()
	t.Heal()
	t.Heal()
	println(t.Markdown())

	println("\n-- WAIT 20 minutes\n")
	os.Sleep(20 * time.Minute)
	println(t.Markdown())

	println("\n-- WAIT 20 hours\n")
	os.Sleep(20 * time.Hour)
	println(t.Markdown())

	println("\n-- WAIT 20 hours\n")
	os.Sleep(20 * time.Hour)
	println(t.Markdown())
}

// Output:
//
// -- INITIAL
//
// # Gnome 😃
//
// * age: 0
// * hunger: 50
// * happiness: 50
// * health: 50
// * sleepy: 0
//
// -- WAIT 20 minutes
//
// # Gnome 😃
//
// * age: 0
// * hunger: 70
// * happiness: 30
// * health: 30
// * sleepy: 20
//
// -- FEEDx3, PLAYx2, HEALx4
//
// # Gnome 😃
//
// * age: 0
// * hunger: 40
// * happiness: 50
// * health: 70
// * sleepy: 20
//
// -- WAIT 20 minutes
//
// # Gnome 😃
//
// * age: 0
// * hunger: 60
// * happiness: 30
// * health: 50
// * sleepy: 40
//
// -- WAIT 20 hours
//
// # Gnome 😵
//
// * age: 20
// * hunger: 0
// * happiness: 0
// * health: 0
// * sleepy: 0
//
// -- WAIT 20 hours
//
// # Gnome 😵
//
// * age: 20
// * hunger: 0
// * happiness: 0
// * health: 0
// * sleepy: 0
