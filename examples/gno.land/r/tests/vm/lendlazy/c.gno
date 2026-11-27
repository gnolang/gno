package lendlazy

import "gno.land/r/tests/vm/lendboth"

// Both and Neither reach lendboth only through this import, so after a restart
// the first call here loads lendboth lazily.
func Both(cur realm) int { return lendboth.Both(cross(cur)) }

func Neither(cur realm) (bool, bool, int) { return lendboth.Neither(cross(cur)) }
