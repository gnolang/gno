package home

import (
	"gno.land/p/moonia/dao"
)

func SetAdmin() string {
	return ds.DAO.SetAdmin()
}

func JoinDAO() string {
	return ds.DAO.JoinDAO()
}

func LeaveDAO() string {
	return ds.DAO.LeaveDAO()
}

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