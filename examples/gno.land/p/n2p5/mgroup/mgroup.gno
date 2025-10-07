// Package mgroup is a simple managed group managing ownership and membership
// for authorization in gno realms. The ManagedGroup struct is used to manage
// the owner, backup owners, and members of a group. The owner is the primary
// owner of the group and can add and remove backup owners and members. Backup
// owners can claim ownership of the group. This is meant to provide backup
// accounts for the owner in case the owner account is lost or compromised.
// Members are used to authorize actions across realms.
package mgroup

import (
	"chain/runtime"
	"errors"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
)

var (
	ErrCannotRemoveOwner = errors.New("mgroup: cannot remove owner")
	ErrNotBackupOwner    = errors.New("mgroup: not a backup owner")
	ErrNotMember         = errors.New("mgroup: not a member")
	ErrInvalidAddress    = errors.New("mgroup: address is invalid")
)

type ManagedGroup struct {
	owner        *ownable.Ownable
	backupOwners *avl.Tree
	members      *avl.Tree
}

// New creates a new ManagedGroup with the owner set to the provided address.
// The owner is automatically added as a backup owner and member of the group.
func New(ownerAddress address) *ManagedGroup {
	g := &ManagedGroup{
		owner:        ownable.NewWithAddress(ownerAddress),
		backupOwners: avl.NewTree(),
		members:      avl.NewTree(),
	}
	err := g.addBackupOwner(ownerAddress)
	if err != nil {
		panic(err)
	}
	err = g.addMember(ownerAddress)
	if err != nil {
		panic(err)
	}
	return g
}

// AddBackupOwner adds a backup owner to the group by address.
// If the caller is not the owner, an error is returned.
func (g *ManagedGroup) AddBackupOwner(addr address) error {
	if !g.owner.OwnedByPrevious() {
		return ownable.ErrUnauthorized
	}
	return g.addBackupOwner(addr)
}

func (g *ManagedGroup) addBackupOwner(addr address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	g.backupOwners.Set(addr.String(), struct{}{})
	return nil
}

// RemoveBackupOwner removes a backup owner from the group by address.
// The owner cannot be removed. If the caller is not the owner, an error is returned.
func (g *ManagedGroup) RemoveBackupOwner(addr address) error {
	if !g.owner.OwnedByPrevious() {
		return ownable.ErrUnauthorized
	}
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	if addr == g.Owner() {
		return ErrCannotRemoveOwner
	}
	g.backupOwners.Remove(addr.String())
	return nil
}

// ClaimOwnership allows a backup owner to claim ownership of the group.
// If the caller is not a backup owner, an error is returned.
// The caller is automatically added as a member of the group.
func (g *ManagedGroup) ClaimOwnership() error {
	caller := runtime.PreviousRealm().Address()
	// already owner, skip
	if caller == g.Owner() {
		return nil
	}
	if !g.IsBackupOwner(caller) {
		return ErrNotMember
	}
	g.owner = ownable.NewWithAddress(caller)
	err := g.addMember(caller)
	return err
}

// AddMember adds a member to the group by address.
// If the caller is not the owner, an error is returned.
func (g *ManagedGroup) AddMember(addr address) error {
	if !g.owner.OwnedByPrevious() {
		return ownable.ErrUnauthorized
	}
	return g.addMember(addr)
}

func (g *ManagedGroup) addMember(addr address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	g.members.Set(addr.String(), struct{}{})
	return nil
}

// RemoveMember removes a member from the group by address.
// The owner cannot be removed. If the caller is not the owner,
// an error is returned.
func (g *ManagedGroup) RemoveMember(addr address) error {
	if !g.owner.OwnedByPrevious() {
		return ownable.ErrUnauthorized
	}
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	if addr == g.Owner() {
		return ErrCannotRemoveOwner
	}
	g.members.Remove(addr.String())
	return nil
}

// MemberCount returns the number of members in the group.
func (g *ManagedGroup) MemberCount() int {
	return g.members.Size()
}

// BackupOwnerCount returns the number of backup owners in the group.
func (g *ManagedGroup) BackupOwnerCount() int {
	return g.backupOwners.Size()
}

// IsMember checks if an address is a member of the group.
func (g *ManagedGroup) IsMember(addr address) bool {
	return g.members.Has(addr.String())
}

// IsBackupOwner checks if an address is a backup owner in the group.
func (g *ManagedGroup) IsBackupOwner(addr address) bool {
	return g.backupOwners.Has(addr.String())
}

// Owner returns the owner of the group.
func (g *ManagedGroup) Owner() address {
	return g.owner.Owner()
}

// BackupOwners returns a slice of all backup owners in the group, using the underlying
// avl.Tree to iterate over the backup owners. If you have a large group, you may
// want to use BackupOwnersWithOffset to iterate over backup owners in chunks.
func (g *ManagedGroup) BackupOwners() []string {
	return g.BackupOwnersWithOffset(0, g.BackupOwnerCount())
}

// Members returns a slice of all members in the group, using the underlying
// avl.Tree to iterate over the members. If you have a large group, you may
// want to use MembersWithOffset to iterate over members in chunks.
func (g *ManagedGroup) Members() []string {
	return g.MembersWithOffset(0, g.MemberCount())
}

// BackupOwnersWithOffset returns a slice of backup owners in the group, using the underlying
// avl.Tree to iterate over the backup owners. The offset and count parameters allow you
// to iterate over backup owners in chunks to support patterns such as pagination.
func (g *ManagedGroup) BackupOwnersWithOffset(offset, count int) []string {
	return sliceWithOffset(g.backupOwners, offset, count)
}

// MembersWithOffset returns a slice of members in the group, using the underlying
// avl.Tree to iterate over the members. The offset and count parameters allow you
// to iterate over members in chunks to support patterns such as pagination.
func (g *ManagedGroup) MembersWithOffset(offset, count int) []string {
	return sliceWithOffset(g.members, offset, count)
}

// sliceWithOffset is a helper function to iterate over an avl.Tree with an offset and count.
func sliceWithOffset(t *avl.Tree, offset, count int) []string {
	var result []string
	t.IterateByOffset(offset, count, func(k string, _ any) bool {
		if k == "" {
			return true
		}
		result = append(result, k)
		return false
	})
	return result
}
