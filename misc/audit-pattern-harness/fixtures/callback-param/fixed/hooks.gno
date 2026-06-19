package hooks

var value int

func Update(cur realm, next int) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	value = next
}

func Value() int {
	return value
}
