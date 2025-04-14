package home

import (
	"std"

	"gno.land/p/moonia/dao"
)

// Admin Methods //

func IsAdmin(addr std.Address) bool {
	return ds.DAO.IsAdmin(addr)
}

func SetAdmin() string {
	return ds.DAO.SetAdmin()
}

func TransferAdmin(addr std.Address) string {
	return ds.DAO.TransferAdmin(addr)
}

func KickMember(addr std.Address) string {
	return ds.DAO.KickMember(addr)
}

func AddMember(addr std.Address) string {
	return ds.DAO.AddMember(addr)
}

func ListMembers() map[std.Address]bool {
	return ds.DAO.Whitelist
}

func AcceptRequest(addr std.Address) string {
	return ds.DAO.AcceptRequest(addr)
}

func DeclineRequest(addr std.Address) string {
	return ds.DAO.DeclineRequest(addr)
}

func ListRequests() map[std.Address]bool {
	return ds.DAO.ListRequests()
}

// DAO Methods //

func CreateDAO(id, name, desc string) string {
	if daoMap[id] != nil {
		panic("DAO with this ID already exists.")
	}
	newDAO := dao.NewDAO(name, desc)
	daoMap[id] = newDAO
	proposalMap[id] = dao.NewProposalStore(newDAO)

	newDAO.SetAdmin()

	return "DAO '" + id + "' created with name: " + name
}

func RequestDAO(daoID string) string {
	d := daoMap[daoID]
	if d == nil {
		panic("DAO not found: " + daoID)
	}
	ds.DAO = d
	return ds.DAO.RequestDAO()
}

func LeaveDAO() string {
	return ds.DAO.LeaveDAO()
}

// Proposals Methods //

func CreateProposal(title, description string, period int64) string {
	return ds.Proposals.CreateProposal(title, description, period)
}

func Vote(indexStr, voteYesStr string) string {
	return ds.Proposals.Vote(indexStr, voteYesStr)
}

func CloseProposal(indexStr string) string {
	return ds.Proposals.CloseProposal(indexStr)
}

func EditProposal(indexStr, newTitle, newDescription string, newPeriod int64) string {
	return ds.Proposals.EditProposal(indexStr, newTitle, newDescription, newPeriod)
}

func GetProposal(indexStr string) dao.Proposal {
	return ds.Proposals.GetProposal(indexStr)
}
