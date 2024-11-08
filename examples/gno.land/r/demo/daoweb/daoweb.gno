package daoweb

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/json"
	"gno.land/r/gov/dao/bridge"
)

// Proposals returns the paginated GovDAO proposals
func Proposals(offset, count uint64) string {
	var (
		propStore = bridge.GovDAO().GetPropStore()
		size      = propStore.Size()
	)

	// Get the props
	props := propStore.Proposals(offset, count)

	resp := ProposalsResponse{
		Proposals: make([]Proposal, 0, count),
		Total:     uint64(size),
	}

	for _, p := range props {
		prop := Proposal{
			Author:      p.Author(),
			Description: p.Description(),
			Status:      p.Status(),
			Stats:       p.Stats(),
			IsExpired:   p.IsExpired(),
		}

		resp.Proposals = append(resp.Proposals, prop)
	}

	// Encode the response into JSON
	encodedProps, err := json.Marshal(encodeProposalsResponse(resp))
	if err != nil {
		panic(err)
	}

	return string(encodedProps)
}

// ProposalByID fetches the proposal using the given ID
func ProposalByID(id uint64) string {
	propStore := bridge.GovDAO().GetPropStore()

	p, err := propStore.ProposalByID(id)
	if err != nil {
		panic(err)
	}

	// Encode the response into JSON
	prop := Proposal{
		Author:      p.Author(),
		Description: p.Description(),
		Status:      p.Status(),
		Stats:       p.Stats(),
		IsExpired:   p.IsExpired(),
	}

	encodedProp, err := json.Marshal(encodeProposal(prop))
	if err != nil {
		panic(err)
	}

	return string(encodedProp)
}

// encodeProposal encodes a proposal into a json node
func encodeProposal(p Proposal) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"author":      json.StringNode("author", p.Author.String()),
		"description": json.StringNode("description", p.Description),
		"status":      json.StringNode("status", p.Status.String()),
		"stats": json.ObjectNode("stats", map[string]*json.Node{
			"yay_votes":          json.NumberNode("yay_votes", float64(p.Stats.YayVotes)),
			"nay_votes":          json.NumberNode("nay_votes", float64(p.Stats.NayVotes)),
			"abstain_votes":      json.NumberNode("abstain_votes", float64(p.Stats.AbstainVotes)),
			"total_voting_power": json.NumberNode("total_voting_power", float64(p.Stats.TotalVotingPower)),
		}),
		"is_expired": json.BoolNode("is_expired", p.IsExpired),
	})
}

// encodeProposalsResponse encodes a proposal response into a JSON node
func encodeProposalsResponse(props ProposalsResponse) *json.Node {
	proposals := make([]*json.Node, 0, len(props.Proposals))

	for _, p := range props.Proposals {
		proposals = append(proposals, encodeProposal(p))
	}

	return json.ObjectNode("", map[string]*json.Node{
		"proposals": json.ArrayNode("proposals", proposals),
		"total":     json.NumberNode("total", float64(props.Total)),
	})
}

// ProposalsResponse is a paginated proposal response
type ProposalsResponse struct {
	Proposals []Proposal `json:"proposals"`
	Total     uint64     `json:"total"`
}

// Proposal is a single GovDAO proposal
type Proposal struct {
	Author      std.Address        `json:"author"`
	Description string             `json:"description"`
	Status      dao.ProposalStatus `json:"status"`
	Stats       dao.Stats          `json:"stats"`
	IsExpired   bool               `json:"is_expired"`
}
