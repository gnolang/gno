package home

import (
	"gno.land/p/demo/ufmt"
)

var count int

// this here serves to increment my counter
func Increment() {
	count++
}

func Decrement() {
	count--
}

func init() {
	count = 0 // arbitrary value
}

func Render(_ string) string {
	out := "# Hey !\n"
	out += "My name is Nicolas, french student in computer science !\n\n"
	out += "Want to hear about me ?\n\n"
	out += RenderPassion()
	out += RenderCounter()
	return out
}

func RenderPassion() string {
	out := "## Passions\n\n"
	out += "### I love:\n\n"
	out += "Crying about Hollow knight silksong never coming out\n\n"
	out += "![Silksong for 2025 confirmed](https://media.discordapp.net/attachments/1221580738487390268/1357382383166291978/FKcRXNtkuNwC7bL24mgznJq5.png?ex=67f0005b&is=67eeaedb&hm=1610610eff0fbbe10c0492bd63ae6c4f57c2c2586c1e52639330b1d394381033&=&format=webp&quality=lossless&width=233&height=350)"
	out += "\n\n"
	out += "Playing the award winning MMORPG Final Fantasy XIV online\n\n"
	out += "And when I'm not doing any of the above, I like coding in C/C++ ~( °v°)\n"
	return out
}

func RenderCounter() string {
	out := "## Secret counter\n\n"
	out += "No one really knows what this counter actually counts\n\n"
	out += "**" + ufmt.Sprintf("%d", count) + "**"
	out += "\n\n\n... but it really ain't that high\n\n"
	return out
}
