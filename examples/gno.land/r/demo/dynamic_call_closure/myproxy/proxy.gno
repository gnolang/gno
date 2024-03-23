package myproxy

// no import needed.

func ExecAsMe(callback func()) {
	assertConditionsAreMet()
	callback()
}

func assertConditionsAreMet() {
	// here we can make assertions for security.
	// std.GetOrigCaller()
	// std.PrevRealm()
	// std.GetCallerAt(2)
	// ...
}
