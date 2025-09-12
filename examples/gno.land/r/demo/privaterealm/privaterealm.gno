package privaterealm

// this realm is an example of a private realm
// it cannot be imported and other realms cannot store ref to objects in this realm or object that use type from this realm

func Render(path string) string {
	return "private"
}
