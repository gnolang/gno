package tamagotchi

import (
	"chain/runtime"

	"gno.land/p/demo/tamagotchi"
	"gno.land/p/nt/ufmt"
)

var t *tamagotchi.Tamagotchi

func init(cur realm) {
	Reset(cur, "gnome#0")
}

func Reset(cur realm, optionalName string) string {
	name := optionalName
	if name == "" {
		height := runtime.ChainHeight()
		name = ufmt.Sprintf("gnome#%d", height)
	}

	t = tamagotchi.New(name)

	return ufmt.Sprintf("A new tamagotchi is born. Their name is %s %s.", t.Name(), t.Face())
}

func Feed(cur realm) string {
	t.Feed()
	return t.Markdown()
}

func Play(cur realm) string {
	t.Play()
	return t.Markdown()
}

func Heal(cur realm) string {
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
