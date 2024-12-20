package dao_realm

import (
	"fmt"
	"testing"

	"gno.land/p/demo/json"
	"gno.land/p/teritori/dao_voting_group"
	"gno.land/p/teritori/havl"
)

func TestInit(t *testing.T) {
	{
		proposalsJSON := getProposalsJSON(0, 42, "TODO", false)
		expected := `[]`
		if proposalsJSON != expected {
			t.Fatalf("Expected %s, got %s", expected, proposalsJSON)
		}
	}

	{
		members := group.GetMembers("", "", 0, havl.Latest)
		iSlice := make([]*json.Node, len(members))
		for i, v := range members {
			iSlice[i] = v.ToJSON()
		}

		membersJSON := json.ArrayNode("", iSlice).String()
		expected := `[{"address":"g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv","power":"1"},{"address":"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a","power":"1"},{"address":"g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c","power":"1"},{"address":"g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym","power":"1"},{"address":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","power":"1"}]`
		if membersJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, membersJSON)
		}
	}
}

func TestUpdateMembers(t *testing.T) {
	var membersJSON string

	{
		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": [{"type": "gno.land/p/teritori/dao_voting_group.UpdateMembers", "payload": [{"power": "2", "address": "g18syxa0vh0vmne90mwhtynjet0zgeqf6prh3ryy"}]}]}`))
		VoteJSON(0, id, `{"vote": "Yes", "rationale": "testing"}`)
		Execute(0, id)
		members := group.GetMembers("", "", 0, havl.Latest)
		iSlice := make([]*json.Node, len(members))
		for i, v := range members {
			iSlice[i] = v.ToJSON()
		}

		membersJSON = json.ArrayNode("", iSlice).String()
		expected := fmt.Sprintf(`[{"address":"g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv","power":"1"},{"address":"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a","power":"1"},{"address":"g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c","power":"1"},{"address":"g18syxa0vh0vmne90mwhtynjet0zgeqf6prh3ryy","power":"2"},{"address":"g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym","power":"1"},{"address":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","power":"1"}]`)
		if membersJSON != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, membersJSON)
		}

		totalPower := group.TotalPowerAtHeight(havl.Latest, nil)
		if totalPower != 7 {
			t.Errorf("Expected total power to be 6, got %d", totalPower)
		}
	}

	{
		children := json.Must(json.Unmarshal([]byte(membersJSON))).MustArray()
		if len(children) != 6 {
			t.Errorf("Expected 6 members, got %d", len(children))
		}

		var member dao_voting_group.Member
		member.FromJSON(children[0])

		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop 2", "description": "A description", "messages": [{"type": "gno.land/p/teritori/dao_voting_group.UpdateMembers", "payload": [{"address": "%s", "power": "0"}]}]}`, member.Address.String()))
		VoteJSON(0, id, `{"vote": "Yes", "rationale": "testing"}`)
		Execute(0, id)
		members := group.GetMembers("", "", 0, havl.Latest)
		iSlice := make([]*json.Node, len(members))
		for i, v := range members {
			iSlice[i] = v.ToJSON()
		}

		membersJSON := json.ArrayNode("", iSlice).String()
		expected := `[{"address":"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a","power":"1"},{"address":"g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c","power":"1"},{"address":"g18syxa0vh0vmne90mwhtynjet0zgeqf6prh3ryy","power":"2"},{"address":"g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym","power":"1"},{"address":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","power":"1"}]`
		if membersJSON != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, membersJSON)
		}

		totalPower := group.TotalPowerAtHeight(havl.Latest, nil)
		if totalPower != 6 {
			t.Errorf("Expected total power to be 6, got %d", totalPower)
		}
	}
}

func TestUpdateSettings(t *testing.T) {
	// not sure why but in this test the proposal ids start at 3 and the voting power is 5 when all tests are run, shouldn't tests be isolated? TODO: investigate

	{
		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": [{"type": "gno.land/p/teritori/dao_proposal_single.UpdateSettings", "payload": {"threshold": {"thresholdQuorum": {"threshold": {"percent": 200}, "quorum": {"percent": 200}}}}}]}`))
		VoteJSON(0, id, `{"vote": "Yes", "rationale": "testing"}`)
		Execute(0, id)
		proposalJSON := getProposalJSON(0, id)
		expected := fmt.Sprintf(`{"id":"%d","title":"Test prop","description":"A description","proposer":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","startHeight":"123","totalPower":"6","messages":[{"type":"gno.land/p/teritori/dao_proposal_single.UpdateSettings","payload":{"threshold":{"thresholdQuorum":{"threshold":{"percent":200},"quorum":{"percent":200}}}}}],"status":"Executed","votes":{"yes":"1","no":"0","abstain":"0"},"allowRevoting":false,"ballots":{"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm":{"power":"1","vote":"Yes","rationale":"testing"}},"expiration":{"atTime":"2009-03-27T23:31:30Z"},"threshold":{"thresholdQuorum":{"threshold":{"percent":100},"quorum":{"percent":100}}}}`, id)
		if proposalJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, proposalJSON)
		}
	}

	{
		// make sentiment proposal
		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": []}`))
		VoteJSON(0, id, `{"vote": "Yes", "rationale": "testing"}`)
		Execute(0, id)
		proposalJSON := getProposalJSON(0, id)
		expected := fmt.Sprintf(`{"id":"%d","title":"Test prop","description":"A description","proposer":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","startHeight":"123","totalPower":"6","messages":[],"status":"Executed","votes":{"yes":"1","no":"0","abstain":"0"},"allowRevoting":false,"ballots":{"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm":{"power":"1","vote":"Yes","rationale":"testing"}},"expiration":{"atTime":"2009-03-27T23:31:30Z"},"threshold":{"thresholdQuorum":{"threshold":{"percent":200},"quorum":{"percent":200}}}}`, id)
		if proposalJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, proposalJSON)
		}
	}
}
