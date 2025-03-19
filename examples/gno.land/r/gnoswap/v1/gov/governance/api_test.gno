package governance

import (
	"strconv"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestCreateProposalJsonNode(t *testing.T) {
	propAddr := testutils.TestAddress("proposal")
	title := "Test Proposal"
	desc := "This is a test proposal"

	testProposal := ProposalInfo{
		ConfigVersion: 1,
		Proposer:      propAddr,
		ProposalType:  Text,
		Title:         title,
		Description:   desc,
	}

	result := createProposalJsonNode(123, testProposal)

	expectedFields := []struct {
		key      string
		expected string
	}{
		{"id", "123"},
		{"configVersion", "1"},
		{"proposer", propAddr.String()},
		{"type", Text.String()},
		{"title", title},
		{"description", desc},
	}

	for _, field := range expectedFields {
		node, err := result.GetKey(field.key)
		uassert.NoError(t, err)

		value, err := node.GetString()
		uassert.NoError(t, err)
		uassert.Equal(t, value, field.expected)
	}

	encodedFields := []string{"status", "vote", "extra"}
	for _, field := range encodedFields {
		node, err := result.GetKey(field)
		if err != nil {
			t.Errorf("field not found: %s", field)
			continue
		}

		value, err := node.GetString()
		if err != nil {
			t.Errorf("failed to get value: %s", err)
			continue
		}

		decodedValue := b64Decode(value)
	}
}

func TestCreateProposalJsonNode_CheckRequiredFields(t *testing.T) {
	emptyProposal := ProposalInfo{
		ConfigVersion: 0,
		Proposer:      testutils.TestAddress("proposal"),
		ProposalType:  Text,
		Title:         "",
		Description:   "",
	}

	result := createProposalJsonNode(0, emptyProposal)

	requiredFields := []string{
		"id", "configVersion", "proposer", "status",
		"type", "title", "description", "vote", "extra",
	}

	for _, field := range requiredFields {
		_, err := result.GetKey(field)
		uassert.NoError(t, err)
	}
}

func TestGetProposalStatus(t *testing.T) {
	proposals = avl.NewTree()

	// Test Case 1: Non-existent proposal
	status := getProposalStatus(999)
	uassert.Equal(t, status, "")

	// Test Case 2: Active proposal
	now := uint64(time.Now().Unix())
	proposal := ProposalInfo{
		ConfigVersion: 1,
		State: ProposalState{
			CreatedAt:  now,
			Upcoming:   false,
			Active:     true,
			Passed:     false,
			PassedAt:   0,
			Rejected:   false,
			RejectedAt: 0,
			Canceled:   false,
			CanceledAt: 0,
			Executed:   false,
			ExecutedAt: 0,
			Expired:    false,
			ExpiredAt:  0,
		},
	}

	proposals.Set("1", proposal)

	status = getProposalStatus(1)

	node, err := json.Unmarshal([]byte(status))
	uassert.NoError(t, err)
	uassert.True(t, node.IsObject())

	tests := []struct {
		key      string
		expected string
	}{
		{"createdAt", strconv.FormatUint(now, 10)},
		{"upcoming", "false"},
		{"active", "true"},
		{"passed", "false"},
		{"passedAt", "0"},
		{"rejected", "false"},
		{"rejectedAt", "0"},
		{"canceled", "false"},
		{"canceledAt", "0"},
		{"executed", "false"},
		{"executedAt", "0"},
		{"expired", "false"},
		{"expiredAt", "0"},
	}

	for _, tc := range tests {
		uassert.True(t, node.HasKey(tc.key))

		value, err := node.GetKey(tc.key)
		uassert.NoError(t, err)
		uassert.True(t, value.IsString())

		str, err := value.GetString()
		uassert.NoError(t, err)
		uassert.Equal(t, str, tc.expected)
	}

	// Test Case 3: State transition
	proposal.State.Active = false
	proposal.State.Passed = true
	proposal.State.PassedAt = now + 500
	proposal.State.Executed = true
	proposal.State.ExecutedAt = now + 1000

	proposals.Set("2", proposal)

	status = getProposalStatus(2)
	node2, err := json.Unmarshal([]byte(status))
	uassert.NoError(t, err)

	// Test node traversal using ObjectEach
	expectedFields := map[string]string{
		"active":     "false",
		"passed":     "true",
		"passedAt":   strconv.FormatUint(now+500, 10),
		"executed":   "true",
		"executedAt": strconv.FormatUint(now+1000, 10),
	}

	node2.ObjectEach(func(key string, value *json.Node) {
		if expected, ok := expectedFields[key]; ok {
			str, err := value.GetString()
			uassert.NoError(t, err)
			uassert.Equal(t, str, expected)
		}
	})
}
