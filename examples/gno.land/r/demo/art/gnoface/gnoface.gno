package gnoface

import (
	"math/rand"
	"strconv"
	"strings"

	"gno.land/p/demo/entropy"
	"gno.land/p/nt/ufmt/v0"
)

func Render(path string) string {
	seed := uint64(entropy.New().Value())

	path = strings.TrimSpace(path)
	if path != "" {
		s, err := strconv.Atoi(path)
		if err != nil {
			panic(err)
		}
		seed = uint64(s)
	}

	output := ufmt.Sprintf("Gnoface #%d\n", seed)
	output += "```\n" + Draw(seed) + "```\n"
	return output
}

func Draw(seed uint64) string {
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

	r := rand.New(rand.NewPCG(seed, 0xdeadbeef))

	return pick(r, hairs) + "\n" +
		pick(r, headtop) + "\n" +
		pick(r, headspace) + "\n" +
		" | " + pick(r, eyebrow) + "   " + pick(r, eyebrow) + " |\n" +
		pick(r, ear) + pick(r, eyesmiddle) + pick(r, ear) + "\n" +
		pick(r, headspace) + "\n" +
		pick(r, nose) + "\n" +
		pick(r, headspace) + "\n" +
		pick(r, mouth) + "\n" +
		pick(r, headspace) + "\n" +
		pick(r, headbottom) + "\n"
}

func pick(r *rand.Rand, slice []string) string {
	return slice[r.IntN(len(slice))]
}

// based on https://github.com/moul/pipotron/blob/master/dict/ascii-face.yml
