// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import "strings"

// Subpath validation grammar: '/'-separated segments of [a-z0-9], with
// '_.-' allowed only inside a segment. try is non-crossing (the `_ int`
// discriminator) so main's cur stays the topmost crossing cur.
func try(_ int, rlm realm, sp string) {
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
	try(0, cur, "")                        // empty
	try(0, cur, "Dao")                     // uppercase
	try(0, cur, "a b")                     // space
	try(0, cur, "a:b")                     // colon
	try(0, cur, "dao/")                    // trailing slash
	try(0, cur, "/dao")                    // leading slash
	try(0, cur, "a//b")                    // empty segment
	try(0, cur, "..")                      // traversal segment
	try(0, cur, "../x")                    // traversal
	try(0, cur, "_x")                      // edge punctuation
	try(0, cur, "x-")                      // edge punctuation
	try(0, cur, strings.Repeat("a", 256))  // synthesized > 256 (host+':'+256)

	// Accepted.
	try(0, cur, "dao/42")
	try(0, cur, "v1.2")
	try(0, cur, "role_admin")
	try(0, cur, "a")

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
