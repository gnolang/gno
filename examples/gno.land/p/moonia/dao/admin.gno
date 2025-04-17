package dao

import (
	"std"
)

// Admin Methods //

func (d *DAO) TransferAdmin(addr std.Address) string {
	caller := std.PreviousRealm().Address()

	if !d.IsAdmin(caller) {
		panic("Only the admin can transfer admin.")
	}
	if d.IsAdmin(addr) {
		panic("Address is already the admin.")
	}
	if !d.IsMember(addr) {
		panic("New admin must be a member.")
	}
	d.Admin = addr
	return "Admin transferred to " + addr.String()
}

func (d *DAO) SetAdmin() string {
	if d.Admin != "" {
		panic("Admin is already set.")
	}
	caller := std.PreviousRealm().Address()
	d.Admin = caller
	d.Whitelist[caller] = true
	return "Admin set to: " + caller.String()
}

func (d *DAO) IsAdmin(addr std.Address) bool {
	return addr == d.Admin
}

func (d *DAO) AcceptRequest(addr std.Address) string {
	caller := std.PreviousRealm().Address()
	if !d.IsAdmin(caller) {
		panic("Only admin can accept requests.")
	}
	if !d.Requests[addr] {
		panic("No such request.")
	}
	delete(d.Requests, addr)
	return d.AddMember(addr)
}

func (d *DAO) DeclineRequest(addr std.Address) string {
	caller := std.PreviousRealm().Address()
	if !d.IsAdmin(caller) {
		panic("Only admin can decline requests.")
	}
	if !d.Requests[addr] {
		panic("No such request.")
	}
	delete(d.Requests, addr)
	return "Request declined for " + addr.String()
}

// Member Methods //

func (d *DAO) IsMember(addr std.Address) bool {
	return d.Whitelist[addr]
}

func (d *DAO) AddMember(addr std.Address) string {
	caller := std.PreviousRealm().Address()

	if !d.IsAdmin(caller) {
		panic("Only the admin can add members.")
	}
	if d.IsMember(addr) {
		return "Address is already a member."
	}
	d.Whitelist[addr] = true
	return "Member added: " + addr.String()
}

func (d *DAO) KickMember(addr std.Address) string {
	caller := std.PreviousRealm().Address()

	if !d.IsAdmin(caller) {
		panic("Only the admin can kick members.")
	}
	if !d.IsMember(addr) {
		return "Address is not a member."
	}
	delete(d.Whitelist, addr)
	return "Member kicked: " + addr.String()
}

func (d* DAO) ListMembers() map[std.Address]bool {
	return d.Whitelist
}