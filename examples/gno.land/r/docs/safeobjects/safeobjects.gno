package safeobjects

import (
	"chain/runtime"
)

// ExposedSafeObject methods exposed to the public. However, the methods all
// contain a caller check - making them, and the object itself "safe".
// The methods are only callable by the current admin.
// A prime example of a safe object can be found in "p/nt/ownable".
var ExposedSafeObject = SafeObject{
	privilegedData: "This message can only be set by the admin.",
	admin:          address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"),
}

// SafeObject is an object that contains some privileged data
// This data might be privileged because only certain users
// have the edit rights, or for other reasons.
type SafeObject struct {
	privilegedData string
	admin          address
}

// Set is a function only the admin can call
func (so SafeObject) Set(value string) {
	if runtime.CurrentRealm().Address() != so.admin {
		panic("caller is not authorized")
	}

	so.privilegedData = value
}

// UpdateAdmin is a function only the admin can call
func (so SafeObject) UpdateAdmin(newAdmin address) {
	if runtime.CurrentRealm().Address() != so.admin {
		panic("caller is not authorized")
	}

	if !newAdmin.IsValid() {
		panic("new admin address is invalid")
	}

	so.admin = newAdmin
}

// Get is callable by anyone
func (so SafeObject) Get() string {
	return so.privilegedData
}

func Render(_ string) string {
	return `
# Safe Objects

**Safe Objects** are objects that can be exposed at the realm top level
while still keeping the write access to their memory limited.

Safe Objects allow only authorized users (like admins) can modify their
internal state, even if they are exposed at the realm top level.

A prime example of a commonly used safe object is the Ownable object, found under in [**p/nt/ownable**](/p/nt/ownable).

To call methods on exposed safe objects, users need to use [MsgRun](/r/docs/complexargs).

---

` + "`r/docs/safeobjects.ExposedSafeObject`" + ` values:
- Message: ` + ExposedSafeObject.privilegedData + `
- Admin: ` + ExposedSafeObject.admin.String()

}
