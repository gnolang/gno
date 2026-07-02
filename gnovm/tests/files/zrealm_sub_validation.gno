// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import "strings"

// Subpath validation rules. try is a non-crossing helper (the `_ int`
// discriminator keeps the realm param from making it crossing); the
// strict entry guard still passes because main's cur remains the
// topmost crossing frame's Cur.
func try(_ int, rlm realm, sp string) {
	defer func() {
		println("recovered:", recover())
	}()
	rlm.Sub(sp)
}

func main(cur realm) {
	try(0, cur, "")
	try(0, cur, strings.Repeat("a", 257))
	try(0, cur, "a:b")
	try(0, cur, "a\x00b")
	println("valid still works:", cur.Sub(strings.Repeat("a", 256)).IsCurrent())
}

// Output:
// recovered: Sub: subpath cannot be empty
// recovered: Sub: subpath too long (max 256 bytes)
// recovered: Sub: subpath cannot contain ':'
// recovered: Sub: subpath cannot contain NUL byte
// valid still works: true
