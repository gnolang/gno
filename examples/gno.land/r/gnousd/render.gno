package gnousd

import (
	"std"
	"strings"

	"gno.land/p/ufmt"
)

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		str := "# GnoUSD ($GUSD)\n\n"
		str += ufmt.Sprintf("Coin price: %d\n", controller.coinPrice)
		str += ufmt.Sprintf("Share price: %d\n", controller.sharePrice)
		str += ufmt.Sprintf("Cycles: %d\n", controller.cycles.Size())
		cycle := controller.getCurrentCycle()
		str += ufmt.Sprintf("Height (start, current, max): %d, %d, %d\n",
			cycle.startHeight,
			std.GetHeight(),
			cycle.maxHeight(),
		)
		return str
	case c == 2 && parts[0] == "balance":
		return "todo"
	default:
		return "404\n"
	}
}
