package config

import (
	"gno.land/p/n2p5/mgroup"
	"gno.land/p/nt/ufmt"
)

const (
	originalOwner = "g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t" // n2p5
)

var (
	adminGroup  = mgroup.New(originalOwner)
	description = ""
)

// AddBackupOwner adds a backup owner to the Owner Group.
// A backup owner can claim ownership of the contract.
func AddBackupOwner(cur realm, addr address) {
	if err := adminGroup.AddBackupOwner(addr); err != nil {
		panic(err)
	}
}

// RemoveBackupOwner removes a backup owner from the Owner Group.
// The primary owner cannot be removed.
func RemoveBackupOwner(cur realm, addr address) {
	if err := adminGroup.RemoveBackupOwner(addr); err != nil {
		panic(err)
	}
}

// ClaimOwnership allows an authorized user in the ownerGroup
// to claim ownership of the contract.
func ClaimOwnership(cur realm) {
	if err := adminGroup.ClaimOwnership(); err != nil {
		panic(err)
	}
}

// AddAdmin adds an admin to the Admin Group.
func AddAdmin(cur realm, addr address) {
	if err := adminGroup.AddMember(addr); err != nil {
		panic(err)
	}
}

// RemoveAdmin removes an admin from the Admin Group.
// The primary owner cannot be removed.
func RemoveAdmin(cur realm, addr address) {
	if err := adminGroup.RemoveMember(addr); err != nil {
		panic(err)
	}
}

// Owner returns the current owner of the claims contract.
func Owner() address {
	return adminGroup.Owner()
}

// BackupOwners returns the current backup owners of the claims contract.
func BackupOwners() []string {
	return adminGroup.BackupOwners()
}

// Admins returns the current admin members of the claims contract.
func Admins() []string {
	return adminGroup.Members()
}

// IsAdmin checks if an address is in the config adminGroup.
func IsAdmin(addr address) bool {
	return adminGroup.IsMember(addr)
}

// toMarkdownList formats a slice of strings as a markdown list.
func toMarkdownList(items []string) string {
	var result string
	for _, item := range items {
		result += ufmt.Sprintf("- %s\n", item)
	}
	return result
}

func Render(path string) string {
	owner := adminGroup.Owner().String()
	backupOwners := toMarkdownList(BackupOwners())
	adminMembers := toMarkdownList(Admins())
	return ufmt.Sprintf(`
# Config Dashboard

This dashboard shows the current configuration owner, backup owners, and admin members.
- The owner has the exclusive ability to manage the backup owners and admin members.
- Backup owners can claim ownership of the contract and become the owner.
- Admin members are used to authorize actions in other realms, such as [my home realm](/r/n2p5/home).

#### Owner

%s

#### Backup Owners

%s

#### Admin Members

%s

`,
		owner,
		backupOwners,
		adminMembers)
}
