package t56c

import "gno.land/r/tests/vm/t56d"

// Both and Neither reach t56d only through this import, so after a restart
// the first call here loads t56d lazily.
func Both(cur realm) int { return t56d.Both(cross(cur)) }

func Neither(cur realm) (bool, bool, int) { return t56d.Neither(cross(cur)) }
