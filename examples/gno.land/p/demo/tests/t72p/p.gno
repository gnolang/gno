// Package t72p embeds a t72c type, so t72c's Put is promoted onto P.
package t72p

import "gno.land/p/demo/tests/t72c"

type P struct{ t72c.C }
