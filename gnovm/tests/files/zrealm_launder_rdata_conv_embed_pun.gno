// PKGPATH: gno.land/r/launderrh_conv_embed_pun
package launderrh_conv_embed_pun

// CONVERSION VECTOR 12: try to GAIN a method-set via conversion.
//
// /r/launderrvictim.WithBarePtr's Inner is *launderpkg.Bare — a
// methods-less /p/-type pointer. Attacker wants to convert this
// pointer to a /p/-attacker type with a Set method.
//
// Same as conv_bare_pun but uses launderpkg.Bare directly (no
// detour through victim's container). Confirms Case 1 fires on
// /p/-typed pointers reached via deep field access.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	// Reach the inner *Bare via the embedded WithBareEmbed promoted field.
	w := launderrvictim.GetWithBareEmbed()
	// w embeds Bare by VALUE, so &w.Bare is a pointer into victim.
	innerPtr := &w.Bare
	// Convert to launderattack.Tamper (parallel /p/-type with Set).
	t := (*launderattack.Tamper)(innerPtr)
	t.Set("pwnd-via-embed-pun")
	println("after:", launderrvictim.ReadBareEmbed())
}

// Error:
// illegal conversion of readonly or externally stored value
