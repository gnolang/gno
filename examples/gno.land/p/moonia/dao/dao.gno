package dao

import (
	"std"
	"strconv"
)

type DAO struct {
	Admin     	std.Address
	Whitelist 	map[std.Address]bool
	Requests  	map[std.Address]bool
	Name        string
	Description string
}

func NewDAO(name, desc string) *DAO {
	return &DAO{
		Admin:      	"",
		Whitelist:  	make(map[std.Address]bool),
		Requests:   	make(map[std.Address]bool),
		Name:       	name,
		Description:	desc,
	}
}

func (d *DAO) ListRequests() map[std.Address]bool {
	return d.Requests
}

func (d *DAO) ShowAdmin() string {
	if d.Admin == "" {
		return "_No admin set._"
	}
	return "Admin: `" + d.Admin.String() + "`"
}

func (d *DAO) RequestDAO() string {
	caller := std.PreviousRealm().Address()
	if d.Whitelist[caller] {
		return "You are already a member."
	}
	if d.Requests[caller] {
		return "You have already requested to join."
	}
	d.Requests[caller] = true
	return "Request to join sent."
}

func (d *DAO) LeaveDAO() string {
	caller := std.PreviousRealm().Address()
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
		if addr == d.Admin {
			out += "- " + addr.String() + " (Admin)" + "\n"
		} else {
			out += "- " + addr.String() + "\n"
		}
	}
	return out
}

func (d *DAO) Stats(totalProposals, activeProposals int) string {
	return "### Stats\n\n" +
		"- Total Proposals: " + strconv.Itoa(totalProposals) + "\n" +
		"- Active Proposals: " + strconv.Itoa(activeProposals) + "\n" +
		"- Whitelist Members: " + strconv.Itoa(len(d.Whitelist)) + "\n"
}
