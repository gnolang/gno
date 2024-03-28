package priint

import (
	"std"
)

/*
 * This realm simply prints user address and similar information.
 * It can help designing wallet connection.
 *
 * See also r/demo/echo (to print various thing like user address)
 */
func Render(path string) string {
	s := ""
	// add more below as necessary
	s = s + "* path: " + path + "\n"
	s = s + "* std.PrevRealm().Addr(): " + string(std.PrevRealm().Addr()) + "\n"
	return s
}
