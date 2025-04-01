package crossrealm

// This is a top function that does switch realms.
func SRWithSwitch(cb func() string) string {
	switchrealm()

	return cb()
}

// This is a top function that doesn't switch realms.
func SRWithoutSwitch(cb func() string) string {
	return cb()
}
