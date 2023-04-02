package gnoface

import (
	"strings"

	"gno.land/p/demo/rand"
)

func Draw(seed int64) string {
	var (
		hairs = []string{
			"     s",
			"  .......",
			"   s s s",
			"   /\\ /\\",
			"  |||||||",
		}
		headtop = []string{
			" /-------\\",
			" /~~~~~~~\\",
			" /|||||||\\",
			" ////////\\",
			" |||||||||",
			" /\\\\\\\\\\\\\\\\",
		}
		headspace = []string{
			" |       |",
		}
		eyebrow = []string{
			"~",
			"*",
			"_",
			".",
		}
		ear = []string{
			"o",
			" ",
			"D",
			"O",
			"<",
			">",
			".",
			"|",
			")",
			"(",
		}
		eyesmiddle = []string{
			"| o   o |",
			"| o   _ |",
			"| _   o |",
			"| .   . |",
			"| O   O |",
			"| v   v |",
			"| X   X |",
			"| x   X |",
			"| X   D |",
			"| ~   ~ |",
		}
		nose = []string{
			" |   o   |",
			" |   O   |",
			" |   V   |",
			" |   L   |",
			" |   C   |",
			" |   ~   |",
			" |  . .  |",
			" |   .   |",
		}
		mouth = []string{
			" |  __/  |",
			" |  \\_/  |",
			" |   .   |",
			" |  ___  |",
			" |  ~~~  |",
			" |  ===  |",
			" |  <=>  |",
		}
		headbottom = []string{
			" \\-------/",
			" \\~~~~~~~/",
			" \\_______/",
		}
	)

	r := rand.FromSeed(seed)

	output := pick(r, hairs) + "\n" +
		pick(r, headtop) + "\n" +
		pick(r, headspace) + "\n" +
		" | " + pick(r, eyebrow) + "   " + pick(r, eyebrow) + " |\n" +
		pick(r, ear) + pick(r, eyesmiddle) + pick(r, ear) + "\n" +
		pick(r, headspace) + "\n" +
		pick(r, nose) + "\n" +
		pick(r, headspace) + "\n" +
		pick(r, mouth) + "\n" +
		pick(r, headspace) + "\n" +
		pick(r, headbottom)
	return output
}

func pick(r *rand.Instance, slice []string) string {
	return slice[r.Intn(len(slice))]
}

// based on https://github.com/moul/pipotron/blob/master/dict/ascii-face.yml
