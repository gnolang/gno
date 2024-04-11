package groups

import (
	"testing"

	"gno.land/p/demo/json"
)

func TestFromJSON(t *testing.T) {
	{
		addHandler := NewAddMemberHandler()
		j := `{"groupId": "123", "address": "g18syxa0vh0vmne90mwhtynjet0zgeqf6prh3ryy", "weight": "1", "metadata": "hello"}`
		ast := json.Must(json.Unmarshal([]byte(j)))
		imsg := addHandler.Instantiate()
		imsg.FromJSON(ast)
		msg := imsg.(*ExecutableMessageAddMember)
		if msg.GroupID != 123 {
			t.Errorf("Expected group_id to be 123, got %d", msg.GroupID)
		}
		if msg.Address != "g18syxa0vh0vmne90mwhtynjet0zgeqf6prh3ryy" {
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
		j := `{"groupId": "123", "memberId": "42"}`
		ast := json.Must(json.Unmarshal([]byte(j)))
		imsg := deleteHandler.Instantiate()
		imsg.FromJSON(ast)
		msg := imsg.(*ExecutableMessageDeleteMember)
		if msg.GroupID != 123 {
			t.Errorf("Expected group_id to be 123, got %d", msg.GroupID)
		}
		if msg.MemberID != 42 {
			t.Errorf("Expected member_id to be 42, got %d", msg.MemberID)
		}
	}
}
