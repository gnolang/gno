// this package is a joke... or not.
// to be honest, I wish we can improve it so that printfdebugging becomes easier.
package printfdebugging

import "gno.land/p/demo/ufmt"

func BigRedLine() {
	println(ufmt.Sprintf("%s[%dm####################################%s[%dm", Escape, int(BgRed), Escape, int(Reset)))
}
