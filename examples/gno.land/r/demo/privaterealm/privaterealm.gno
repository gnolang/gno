package privaterealm

// ---------------- Private Realm ----------------
//
// This package illustrates a private realm.
//
// Rules of private realms are:
// - Cannot be imported by other realms.
// - References to objects owned by this realm cannot be stored outside it.
// - Data whose type is defined in this realm cannot be retained in other realms.
//
// These rules keep private realms interchangeable.
//
// -----------------------------------------------

func Render(path string) string {
	return "private"
}
