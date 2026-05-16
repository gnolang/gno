// PKGPATH: gno.land/r/demo/tamagotchi/filetests/z0_filetest

package z0_filetest

import (
	"gno.land/r/demo/tamagotchi"
)

func main(cur realm) {
	tamagotchi.Reset(cross2(cur), "tamagnotchi")
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
// * [Feed](/r/demo/tamagotchi$help&func=Feed)
// * [Play](/r/demo/tamagotchi$help&func=Play)
// * [Heal](/r/demo/tamagotchi$help&func=Heal)
// * [Reset](/r/demo/tamagotchi$help&func=Reset)
//
