// this package is a joke... or not.
package printfdebugging

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

func BigRedLine(args ...string) {
	println(ufmt.Sprintf("%s[%dm####################################%s[%dm %s",
		Escape, int(BgRed), Escape, int(Reset),
		strings.Join(args, " "),
	))
}
