package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTendermint_QuorumSuperMajority(t *testing.T) {
	t.Parallel()

	var (
		// Equal voting power map
		votingPowerMap = map[string]uint64{
			"1": 1,
			"2": 1,
			"3": 1,
			"4": 1,
		}

		mockMessages = []*mockMessage{
			{
				getSenderFn: func() []byte {
					return []byte("1")
				},
			},
			{
				getSenderFn: func() []byte {
					return []byte("2")
				},
			},
			{
				getSenderFn: func() []byte {
					return []byte("3")
				},
			},
			{
				getSenderFn: func() []byte {
					return []byte("4")
				},
			},
		}

		mockVerifier = &mockVerifier{
			getTotalVotingPowerFn: func(_ uint64) uint64 {
				return uint64(len(votingPowerMap))
			},
			getSumVotingPowerFn: func(messages []Message) uint64 {
				sum := uint64(0)

				for _, message := range messages {
					sum += votingPowerMap[string(message.GetSender())]
				}

				return sum
			},
		}
	)

	testTable := []struct {
		name               string
		messages           []*mockMessage
		shouldHaveMajority bool
	}{
		{
			"4/4 validators",
			mockMessages,
			true,
		},
		{
			"3/4 validators",
			mockMessages[1:],
			true,
		},
		{
			"2/4 validators",
			mockMessages[2:],
			false,
		},
		{
			"1/4 validators",
			mockMessages[:1],
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tm := NewTendermint(
				mockVerifier,
				nil,
				nil,
				nil,
			)

			convertedMessages := make([]Message, 0, len(testCase.messages))

			for _, mockMessage := range testCase.messages {
				convertedMessages = append(convertedMessages, mockMessage)
			}

			assert.Equal(
				t,
				testCase.shouldHaveMajority,
				tm.hasSuperMajority(convertedMessages),
			)
		})
	}
}

func TestTendermint_QuorumFaultyMajority(t *testing.T) {
	t.Parallel()

	var (
		// Equal voting power map
		votingPowerMap = map[string]uint64{
			"1": 1,
			"2": 1,
			"3": 1,
			"4": 1,
		}

		mockMessages = []*mockMessage{
			{
				getSenderFn: func() []byte {
					return []byte("1")
				},
			},
			{
				getSenderFn: func() []byte {
					return []byte("2")
				},
			},
			{
				getSenderFn: func() []byte {
					return []byte("3")
				},
			},
			{
				getSenderFn: func() []byte {
					return []byte("4")
				},
			},
		}

		mockVerifier = &mockVerifier{
			getTotalVotingPowerFn: func(_ uint64) uint64 {
				return uint64(len(votingPowerMap))
			},
			getSumVotingPowerFn: func(messages []Message) uint64 {
				sum := uint64(0)

				for _, message := range messages {
					sum += votingPowerMap[string(message.GetSender())]
				}

				return sum
			},
		}
	)

	testTable := []struct {
		name               string
		messages           []*mockMessage
		shouldHaveMajority bool
	}{
		{
			"4/4 validators",
			mockMessages,
			true,
		},
		{
			"3/4 validators",
			mockMessages[1:],
			true,
		},
		{
			"2/4 validators",
			mockMessages[2:],
			true,
		},
		{
			"1/4 validators",
			mockMessages[:1],
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tm := NewTendermint(
				mockVerifier,
				nil,
				nil,
				nil,
			)

			convertedMessages := make([]Message, 0, len(testCase.messages))

			for _, mockMessage := range testCase.messages {
				convertedMessages = append(convertedMessages, mockMessage)
			}

			assert.Equal(
				t,
				testCase.shouldHaveMajority,
				tm.hasFaultyMajority(convertedMessages),
			)
		})
	}
}
