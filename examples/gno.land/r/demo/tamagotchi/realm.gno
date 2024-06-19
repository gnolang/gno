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
	name := optionalName
	if name == "" {
		height := std.GetHeight()
		name = ufmt.Sprintf("gnome#%d", height)
	}

	t = tamagotchi.New(name)

	return ufmt.Sprintf("A new tamagotchi is born. Their name is %s %s.", t.Name(), t.Face())
}

func Feed() string {
	t.Feed()
	return t.Markdown()
}

func Play() string {
	t.Play()
	return t.Markdown()
}

func Heal() string {
	t.Heal()
	return t.Markdown()
}

func Render(path string) string {
	tama := t.Markdown()
	links := `Actions:
* [Feed](/r/demo/tamagotchi?help&__func=Feed)
* [Play](/r/demo/tamagotchi?help&__func=Play)
* [Heal](/r/demo/tamagotchi?help&__func=Heal)
* [Reset](/r/demo/tamagotchi?help&__func=Reset)
`

	return tama + "\n\n" + links
}
