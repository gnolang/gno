package social_feeds_dao

import (
	"fmt"
	"std"
	"testing"

	proposal_single "gno.land/p/demo/teritori/daodao/proposal_single"
	social_feeds "gno.land/r/demo/social_feeds"
)

func TestSocialFeedsDAO(t *testing.T) {
	std.TestSetOrigCaller("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	// create a post
	postID := social_feeds.CreatePost(feedID, social_feeds.PostID(0), 0, "Hello, world!")
	_ = social_feeds.GetPost(feedID, postID)

	// ban the post
	proposalID := ProposeJSON(0, fmt.Sprintf(`{"title": "Ban post", "description": "", "messages": [{"type":"gno.land/r/demo/social_feeds.BanPost","payload":{"feedId":%d,"postId":%d,"reason":"spam"}}]}`, feedID, postID))
	VoteJSON(0, proposalID, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
	Execute(0, proposalID)

	// check that the post is deleted
	func() {
		defer func() {
			r := recover()
			expected := "Unable to get post"
			if r == nil {
				t.Errorf("Expected panic while getting post after ban")
			} else if r != expected {
				t.Errorf("Expected panic with message '%s', got '%s'", expected, r)
			}
		}()
		_ = social_feeds.GetPost(feedID, postID)
	}()

	// check that the proposal has the correct shape
	expected := `{"id":0,"title":"Ban post","description":"","proposer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5","startHeight":123,"threshold":{"thresholdQuorum":{"threshold":{"percent":1500},"quorum":{"percent":5000}}},"totalPower":2,"messages":[{"type":"gno.land/r/demo/social_feeds.BanPost","payload":{"feedId":1,"postId":1,"reason":"spam"}}],"status":"Executed","votes":{"yes":1,"no":0,"abstain":0},"allowRevoting":false,"ballots":{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5":{"power":1,"vote":"Yes","rationale":"testing"}}}`
	if proposal := getProposalJSON(0, proposalID); proposal != expected {
		t.Errorf("Expected proposal to be:\n%s\nGot:\n%s", expected, proposal)
	}
}
