// Package t72m imports t72a only; t72a's code makes the interface call.
package t72m

import "gno.land/p/demo/tests/t72a"

// Run has t72a call Put on a t72b value with b.
func Run(b []byte) { t72a.Apply(t72a.NewB(), b) }
