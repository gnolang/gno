package deep

// This can't be called from a MsgCall as it isn't crossing.
func Render(path string) string {
	if path == "" {
		return "it works!"
	} else {
		return "hi " + path
	}
}

// For testing with MsgCall.
func RenderCrossing(cur realm, path string) string {
	if path == "" {
		return "it works!"
	} else {
		return "hi " + path
	}
}
