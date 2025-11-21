package migration

import (
	"gno.land/p/nt/commondao"
	"gno.land/p/onbloc/json"
)

func SerializeDAO(dao *commondao.CommonDAO) *json.Node {
	// TODO: Implement
	return nil
}

func SerializeMemberStorage(s commondao.MemberStorage) *json.Node { // TODO: Expect readonly storage
	// TODO: Implement
	return nil
}

func SerializeProposal(p *commondao.Proposal) *json.Node {
	// TODO: Implement (ignore active proposals)
	return nil
}

func SerializeVotingRecord(r commondao.ReadonlyVotingRecord) *json.Node {
	// TODO: Implement
	return nil
}

func SerializeProposalDefinition(p commondao.ProposalDefinition) *json.Node {
	// TODO: Implement (use a migration.ProposalDefinition type)
	return nil
}
