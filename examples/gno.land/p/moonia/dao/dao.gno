package dao

import (
	"std"
	"strconv"
	"gno.land/p/moonia/utils"
)

type DAO struct {
	Admin     std.Address
	Whitelist map[std.Address]bool
}

func NewDAO() *DAO {
	return &DAO{
		Admin:     "",
		Whitelist: make(map[std.Address]bool),
	}
}

func (d *DAO) SetAdmin() string {
	if d.Admin != "" {
		panic("Admin is already set.")
	}
	caller := utils.GetCaller()
	d.Admin = caller
	d.JoinDAO()
	return "Admin set to: " + caller.String()
}

func (d *DAO) ShowAdmin() string {
	if d.Admin == "" {
		return "_No admin set._"
	}
	return "Admin: `" + d.Admin.String() + "`"
}

func (d *DAO) JoinDAO() string {
	caller := utils.GetCaller()
	if d.Whitelist[caller] {
		panic("You're already a member of the DAO.")
	}
	d.Whitelist[caller] = true
	return "You have successfully joined the DAO."
}

func (d *DAO) LeaveDAO() string {
	caller := utils.GetCaller()
	if !d.Whitelist[caller] {
		panic("You are not a member of the DAO.")
	}
	delete(d.Whitelist, caller)
	return "You have successfully left the DAO."
}

func (d *DAO) ShowWhitelist() string {
	out := "## Whitelist Members ✅\n\n"
	if len(d.Whitelist) == 0 {
		return out + "_Whitelist is empty._\n"
	}
	for addr := range d.Whitelist {
		out += "- " + addr.String() + "\n"
	}
	return out
}

func (d *DAO) Stats(totalProposals, activeProposals int) string {
	return "### Stats\n\n" +
		"- Total Proposals: " + strconv.Itoa(totalProposals) + "\n" +
		"- Active Proposals: " + strconv.Itoa(activeProposals) + "\n" +
		"- Whitelist Members: " + strconv.Itoa(len(d.Whitelist)) + "\n"
}
