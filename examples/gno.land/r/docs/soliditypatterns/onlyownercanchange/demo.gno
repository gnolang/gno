package onlyownercanchange

import "std"

var owner std.Address = ""

func InitOwner(cur realm) {
	if len(owner) != 0 {
		panic("owner already defined")
	}
	owner = std.PreviousRealm().Address()
}

func assertOnlyOwner() {
	if std.PreviousRealm().Address() != owner {
		panic("caller isn't the owner")
	}
}

func ChangeOwner(cur realm, newOwner address) {
	assertOnlyOwner()
	owner = newOwner
}

func RenderDemo() string {
	content := "# DEMO\n"
	content += "Owner address: **" + owner.String() + "**\n\n"
	content += "[InitOwner](onlyownercanchange$help&func=InitOwner)\n\n"
	content += "[ChangeOwner](onlyownercanchange$help&func=ChangeOwner)\n\n"
	return content
}
