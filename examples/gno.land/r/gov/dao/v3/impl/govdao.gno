package impl

import "gno.land/r/gov/dao/proxy"

type GovDAO struct {

}

func (g *GovDAO) PreCreateProposal(r proxy.ProposalRequest)(proxy.Author, error) {
	return nil,nil
}

func (g *GovDAO) PostCreateProposal(r proxy.ProposalRequest, pid proxy.ProposalID) {

}

func (g *GovDAO) VoteOnProposal(r proxy.VoteRequest) error {
	return nil
}

func (g *GovDAO) PreGetProposal(pid proxy.ProposalID) error {
	return nil
}

func (g *GovDAO) PostGetProposal(pid proxy.ProposalID, p *proxy.Proposal) error {
	return nil
}

func (g *GovDAO) Render(string) string {
	return ""
}
