package tamagotchi

import (
	"std"

	"gno.land/p/demo/tamagotchi"
	"gno.land/p/demo/ufmt"
)

var t *tamagotchi.Tamagotchi

func init() {
	Reset("gnome#0")
}

func Reset(optionalName string) string {
	crossing()

	name := optionalName
	if name == "" {
		height := std.ChainHeight()
		name = ufmt.Sprintf("gnome#%d", height)
	}

	t = tamagotchi.New(name)

	return ufmt.Sprintf("A new tamagotchi is born. Their name is %s %s.", t.Name(), t.Face())
}

func Feed() string {
	crossing()

	t.Feed()
	return t.Markdown()
}

func Play() string {
	crossing()

	t.Play()
	return t.Markdown()
}

func Heal() string {
	crossing()

	t.Heal()
	return t.Markdown()
}

func Render(path string) string {
	tama := t.Markdown()
	links := `Actions:
* [Feed](/r/demo/tamagotchi$help&func=Feed)
* [Play](/r/demo/tamagotchi$help&func=Play)
* [Heal](/r/demo/tamagotchi$help&func=Heal)
* [Reset](/r/demo/tamagotchi$help&func=Reset)
`

	return tama + "\n\n" + links
}
