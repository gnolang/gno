package hooks

var afterUpdate func()
var value int

// Vulnerable: stores caller-supplied code that later runs inside this realm.
func SetHook(cur realm, hook func()) {
	afterUpdate = hook
}

func Update(cur realm, next int) {
	value = next
	if afterUpdate != nil {
		afterUpdate()
	}
}

func Value() int {
	return value
}
