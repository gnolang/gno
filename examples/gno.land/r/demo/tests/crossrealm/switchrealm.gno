package crossrealm

// This is a top function that does switch realms.
func ExecSR(cb func() string) string {
	switchrealm()

	return cb()
}

// This is a top function that doesn't switch realms.
func Exec(cb func() string) string {
	return cb()
}
