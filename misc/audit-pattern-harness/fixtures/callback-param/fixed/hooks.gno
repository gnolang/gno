package hooks

var value int

// Fixed: expose a narrow value setter instead of accepting caller-supplied code.
func Update(cur realm, next int) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	value = next
}

func Value() int {
	return value
}
