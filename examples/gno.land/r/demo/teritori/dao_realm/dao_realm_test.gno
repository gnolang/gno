package dao_realm

import (
	"fmt"
	"std"
	"testing"

	dao_core "gno.land/p/demo/teritori/daodao/core"
	dao_interfaces "gno.land/p/demo/teritori/daodao/interfaces"
	proposal_single "gno.land/p/demo/teritori/daodao/proposal_single"
	"gno.land/p/demo/teritori/ujson"
	"gno.land/r/demo/teritori/groups"
	modboards "gno.land/r/demo/modboards"
	"gno.land/r/demo/users"
)

func TestInit(t *testing.T) {
	{
		proposalsJSON := getProposalsJSON(0, 42, "TODO", false)
		expected := "[]"
		if proposalsJSON != expected {
			t.Fatalf("Expected %s, got %s", expected, proposalsJSON)
		}
	}

	{
		membersJSON := ujson.FormatSlice(groups.GetMembers(groupID))
		expected := `[{"id":0,"address":"g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":1,"address":"g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":2,"address":"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":3,"address":"g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":4,"address":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"}]`
		if membersJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, membersJSON)
		}
	}
}

func TestUpdateMembers(t *testing.T) {
	var membersJSON string

	{
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic AddMember: %v", r)
				}
			}()
			id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": [{"type": "gno.land/r/demo/teritori/groups.AddMember", "payload": {"groupId": %d, "weight": 1, "address": "0x42", "metadata": "hello"}}]}`, groupID))
			VoteJSON(0, id, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
			Execute(0, id)
			membersJSON = ujson.FormatSlice(groups.GetMembers(groupID))
			expected := fmt.Sprintf(`[{"id":0,"address":"g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":1,"address":"g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":2,"address":"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":3,"address":"g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":4,"address":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":5,"address":"0x42","weight":1,"metadata":"hello","createdAt":"2009-02-13T23:31:30Z"}]`)
			if membersJSON != expected {
				t.Fatalf("Expected:\n%s\nGot:\n%s", expected, membersJSON)
			}
		}()
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Panic DeleteMember: %v", r)
			}
		}()
		children := ujson.ParseSlice(membersJSON)
		if len(children) != 6 {
			t.Fatalf("Expected 6 members, got %d", len(children))
		}
		var member groups.Member
		children[0].ParseAny(&member)

		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop 2", "description": "A description", "messages": [{"type": "gno.land/r/demo/teritori/groups.DeleteMember", "payload": {"groupId": %d, "memberId": %d}}]}`, groupID, member.GetID()))
		VoteJSON(0, id, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
		Execute(0, id)
		membersJSON = ujson.FormatSlice(groups.GetMembers(groupID))
		expected := fmt.Sprintf(`[{"id":1,"address":"g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":2,"address":"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":3,"address":"g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":4,"address":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","weight":1,"metadata":"","createdAt":"2009-02-13T23:31:30Z"},{"id":5,"address":"0x42","weight":1,"metadata":"hello","createdAt":"2009-02-13T23:31:30Z"}]`)
		if membersJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, membersJSON)
		}
	}()
}

func TestModeration(t *testing.T) {
	std.TestSetOrigSend(std.Coins{{Amount: 100000000, Denom: "ugnot"}}, nil)
	boardID, ok := modboards.GetBoardIDFromName(mainBoardName)
	if !ok {
		t.Fatalf("Board %s not found", mainBoardName)
	}
	modboards.CreateThread(boardID, "Test thread", "Offensive stuff")
	modboards.FlagPost(boardID, 1, 1)
	{
		expected := `\[[post](/r/demo/modboards?help&__func=CreateThread&bid=1&body.type=textarea)]\[[see flags](/r/demo/modboards:dao_realm/flags)]

----------------------------------------
## [Test thread](/r/demo/modboards:dao_realm/1)

Offensive stuff
\- [g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm](/r/users:g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm), [2009-02-13 11:31pm UTC](/r/demo/modboards:dao_realm/1) \[[🚩](/r/demo/modboards?help&__func=FlagPost&boardID=1&threadID=1&postID=1)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies, 1 red flags)

`
		render := modboards.Render(mainBoardName)
		if render != expected {
			t.Fatalf("Expected %s, got `%s`", expected, render)
		}
	}

	id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": [{"type": "gno.land/r/demo/modboards.DeletePost", "payload": {"boardId": %d, "threadId": 1, "postId": 1, "reason": "testing"}}]}`, boardID))
	VoteJSON(0, id, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
	Execute(0, id)
	{
		expected := `\[[post](/r/demo/modboards?help&__func=CreateThread&bid=1&body.type=textarea)]\[[see flags](/r/demo/modboards:dao_realm/flags)]

`
		render := modboards.Render(mainBoardName)
		if render != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s`", expected, render)
		}
	}
}

func TestUpdateSettings(t *testing.T) {
	// why is state preversed between tests??
	{
		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": [{"type": "gno.land/p/demo/teritori/daodao/proposal_single.UpdateSettings", "payload": {"threshold": {"thresholdQuorum": {"threshold": {"percent": 200}, "quorum": {"percent": 200}}}}}]}`))
		VoteJSON(0, id, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
		Execute(0, id)
		proposalJSON := getProposalJSON(0, id)
		expected := `{"id":3,"title":"Test prop","description":"A description","proposer":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","startHeight":123,"threshold":{"thresholdQuorum":{"threshold":{"percent":100},"quorum":{"percent":100}}},"totalPower":5,"messages":[{"type":"gno.land/p/demo/teritori/daodao/proposal_single.UpdateSettings","payload":{"threshold":{"thresholdQuorum":{"threshold":{"percent":200},"quorum":{"percent":200}}}}}],"status":"Executed","votes":{"yes":1,"no":0,"abstain":0},"allowRevoting":false,"ballots":{"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm":{"power":1,"vote":"Yes","rationale":"testing"}}}`
		if proposalJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, proposalJSON)
		}
	}

	{
		// make sentiment proposal
		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Test prop", "description": "A description", "messages": []}`))
		VoteJSON(0, id, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
		Execute(0, id)
		proposalJSON := getProposalJSON(0, id)
		expected := `{"id":4,"title":"Test prop","description":"A description","proposer":"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm","startHeight":123,"threshold":{"thresholdQuorum":{"threshold":{"percent":200},"quorum":{"percent":200}}},"totalPower":5,"messages":[],"status":"Executed","votes":{"yes":1,"no":0,"abstain":0},"allowRevoting":false,"ballots":{"g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm":{"power":1,"vote":"Yes","rationale":"testing"}}}`
		if proposalJSON != expected {
			t.Fatalf("Expected:\n%s\nGot:\n%s", expected, proposalJSON)
		}
	}
}

func TestCreateBoard(t *testing.T) {
	{
		id := ProposeJSON(0, fmt.Sprintf(`{"title": "Create moderated board", "description": "", "messages": [{"type":"gno.land/r/demo/modboards.CreateBoard","payload":{"name":"test_board"}}]}`))
		VoteJSON(0, id, fmt.Sprintf(`{"vote": %d, "rationale": "%s"}`, proposal_single.VoteYes, "testing"))
		Execute(0, id)
		render := modboards.Render("test_board")
		expected := `\[[post](/r/demo/modboards?help&__func=CreateThread&bid=2&body.type=textarea)]\[[see flags](/r/demo/modboards:test_board/flags)]

`
		if render != expected {
			t.Fatalf("Expected:\n`%s`\nGot:\n`%s`", expected, render)
		}
	}
}
