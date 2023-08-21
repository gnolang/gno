package groups

import (
	"testing"

	"gno.land/p/demo/jsonutil_v3"
)

func TestFromJSON(t *testing.T) {
	{
		addHandler := NewAddMemberHandler()
		json := `{"groupId": 123, "address": "gno142", "weight": 1, "metadata": "hello"}`
		ast := jsonutil.TokenizeAndParse(json)
		msg := addHandler.FromJSON(ast).(*ExecutableMessageAddMember)
		if msg.GroupID != 123 {
			t.Errorf("Expected group_id to be 123, got %d", msg.GroupID)
		}
		if msg.Address != "gno142" {
			t.Errorf("Expected address to be '0x42', got %s", msg.Address)
		}
		if msg.Weight != 1 {
			t.Errorf("Expected weight to be 1, got %d", msg.Weight)
		}
		if msg.Metadata != "hello" {
			t.Errorf("Expected metadata to be 'hello', got %s", msg.Metadata)
		}
	}
	{
		deleteHandler := NewDeleteMemberHandler()
		json := `{"groupId": 123, "memberId": 42}`
		ast := jsonutil.TokenizeAndParse(json)
		msg := deleteHandler.FromJSON(ast).(*ExecutableMessageDeleteMember)
		if msg.GroupID != 123 {
			t.Errorf("Expected group_id to be 123, got %d", msg.GroupID)
		}
		if msg.MemberID != 42 {
			t.Errorf("Expected member_id to be 42, got %d", msg.MemberID)
		}
	}
}
