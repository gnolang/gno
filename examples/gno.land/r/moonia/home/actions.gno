package home

import (
	"gno.land/p/moonia/dao"
	"std"
)

// Admin Methods //

func IsAdmin(addr std.Address) bool {
	return ds.DAO.IsAdmin(addr)
}

func SetAdmin() string {
	return ds.DAO.SetAdmin()
}

// func TransferAdmin(addr string) string {
// 	return ds.DAO.TransferAdmin(addr)
// }

func KickMember(addr std.Address) string {
	return ds.DAO.KickMember(addr)
}

func AddMember(addr std.Address) string {
	return ds.DAO.AddMember(addr)
}

func ListMembers() map[std.Address]bool {
	return ds.DAO.Whitelist
}

// DAO Methods //

// TODO:
// func RequestDAO() string {
// 	return ds.DAO.RequestDAO()
// }

func JoinDAO() string {
	return ds.DAO.JoinDAO()
}

func LeaveDAO() string {
	return ds.DAO.LeaveDAO()
}

func IsMember(addr std.Address) bool {
	return ds.DAO.Whitelist[addr]
}

// Proposals Methods //

func CreateProposalTest() string {
	return ds.Proposals.CreateProposalTest()
}

func CreateProposal(title string, description string) string {
	return ds.Proposals.CreateProposal(title, description)
}

func Vote(indexStr string, voteYesStr string) string {
	return ds.Proposals.Vote(indexStr, voteYesStr)
}

func CloseProposal(indexStr string) string {
	return ds.Proposals.CloseProposal(indexStr)
}

func EditProposal(indexStr, newTitle, newDescription string) string {
	return ds.Proposals.EditProposal(indexStr, newTitle, newDescription)
}

func GetProposal(indexStr string) dao.Proposal {
	return ds.Proposals.GetProposal(indexStr)
}

// TODO:
// func SetVotingPeriod(period int64) string {
// 	return ds.Proposals.SetVotingPeriod(period)
// }