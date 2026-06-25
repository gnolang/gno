package ownable

import (
	"chain"
)

var owner address

// InitOwner claims ownership of this realm. It can only be called once:
// the first caller becomes the owner, and any later call panics.
func InitOwner(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	if len(owner) != 0 {
		panic("owner already defined")
	}
	owner = cur.Previous().Address()
}

// ChangeOwner transfers ownership to newOwner. Only the current owner may call it.
func ChangeOwner(cur realm, newOwner address) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	if cur.Previous().Address() != owner {
		panic("caller isn't the owner")
	}
	owner = newOwner
	chain.Emit("OwnershipChange", "newOwner", newOwner.String())
}

func Owner() address {
	return owner
}

func RenderDemo() string {
	content := "## Live Demo\n"
	content += "Owner address: **" + owner.String() + "**\n\n"
	content += "[InitOwner](ownable$help&func=InitOwner)\n\n"
	content += "[ChangeOwner](ownable$help&func=ChangeOwner)\n\n"
	return content
}
