// this package is a joke... or not.
package printfdebugging

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

func BigRedLine(args ...string) {
	println(ufmt.Sprintf("%s[%dm####################################%s[%dm %s",
		Escape, int(BgRed), Escape, int(Reset),
		strings.Join(args, " "),
	))
}

func Success() {
	println("   \033[31mS\033[33mU\033[32mC\033[36mC\033[34mE\033[35mS\033[31mS\033[0m   ")
}
