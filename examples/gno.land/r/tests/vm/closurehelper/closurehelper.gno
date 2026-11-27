// Package closurehelper stores a library closure so that callers from other
// realms run it view-only.
package closurehelper

import "gno.land/p/demo/tests/nestclosure/v0"

var outer = nestclosure.Outer()

func GetOuter() func() func([]int) { return outer }
