package hooks

var afterUpdate func()
var value int

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
