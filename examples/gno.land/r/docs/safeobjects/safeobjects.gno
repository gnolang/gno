package safeobjects

import "std"

const admin = "g125em6arxsnj49vx35f0n0z34putv5ty3376fg5" // @leohhhn

// SafeObject is an object that contains some privileged data
// This data might be privileged because only certain users
// have the edit rights, or for other reasons.
type SafeObject struct {
	privilegedData string
}

// Set is a function only the admin can call
func (so SafeObject) Set(value string) {
	if std.PreviousRealm().Address().String() != admin {
		panic("caller is not authorized")
	}

	so.privilegedData = value
}

// Get is callable by anyone
func (so SafeObject) Get() string {
	return so.privilegedData
}

func Render(path string) string {
	out := "todo"

	return out
}
