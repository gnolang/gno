// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import "strings"

// Subpath validation grammar: '/'-separated segments of [a-z0-9], with
// '_.-' allowed only inside a segment. try is non-crossing (its realm
// trails sp instead of leading it) so main's cur stays the topmost
// crossing cur.
func try(sp string, rlm realm) {
	defer func() {
		if r := recover(); r != nil {
			println(sp, "rejected:", r)
		}
	}()
	rlm.Sub(sp)
	println(sp, "accepted")
}

func main(cur realm) {
	// Rejected.
	try("", cur)                       // empty
	try("Dao", cur)                    // uppercase
	try("a b", cur)                    // space
	try("a:b", cur)                    // colon
	try("dao/", cur)                   // trailing slash
	try("/dao", cur)                   // leading slash
	try("a//b", cur)                   // empty segment
	try("..", cur)                     // traversal segment
	try("../x", cur)                   // traversal
	try("_x", cur)                     // edge punctuation
	try("x-", cur)                     // edge punctuation
	try(strings.Repeat("a", 256), cur) // synthesized > 256 (host+':'+256)

	// Accepted.
	try("dao/42", cur)
	try("v1.2", cur)
	try("role_admin", cur)
	try("a", cur)

	// Subpath() accessor is the canonical "am I a sub" surface.
	sub := cur.Sub("dao/42")
	println("Subpath(sub):", sub.Subpath())
	println("Subpath(cur):", cur.Subpath())
}

// Output:
//  rejected: Sub: subpath cannot be empty
// Dao rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// a b rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// a:b rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// dao/ rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// /dao rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// a//b rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// .. rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// ../x rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// _x rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// x- rejected: Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa rejected: Sub: synthesized pkgpath too long (max 256 bytes)
// dao/42 accepted
// v1.2 accepted
// role_admin accepted
// a accepted
// Subpath(sub): dao/42
// Subpath(cur):
