package dao

import (
	"gno.land/p/moonia/utils"
	"std"
)

// Admin Methods //

func (d *DAO) SetAdmin() string {
	if d.Admin != "" {
		panic("Admin is already set.")
	}
	caller := utils.GetCaller()
	d.Admin = caller
	d.JoinDAO()

	return "Admin set to: " + caller.String()
}

func (d *DAO) IsAdmin(addr std.Address) bool {
	return addr == d.Admin
}

// Member Methods //

func (d *DAO) AddMember(addr std.Address) string {
	caller := utils.GetCaller()

	if !d.IsAdmin(caller) {
		panic("Only the admin can add members.")
	}
	if d.Whitelist[addr] {
		return "Address is already a member."
	}
	d.Whitelist[addr] = true
	return "Member added: " + addr.String()
}

func (d *DAO) KickMember(addr std.Address) string {
	caller := utils.GetCaller()

	if !d.IsAdmin(caller) {
		panic("Only the admin can kick members.")
	}
	if !d.Whitelist[addr] {
		return "Member not found."
	}
	delete(d.Whitelist, addr)
	return "Member kicked: " + addr.String()
}

func (d* DAO) ListMembers() map[std.Address]bool {
	return d.Whitelist
}