package protocol_fee

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestApiGetAccuTransferToGovStaker(t *testing.T) {
	tests := []struct {
		name        string
		historyItem []struct {
			token  string
			amount uint64
		}
		expectedResultAccu string
	}{
		{
			name: "api get accu transfer to gov staker success",
			historyItem: []struct {
				token  string
				amount uint64
			}{
				{token: "gno.land/r/onbloc/bar", amount: 1000},
				{token: "gno.land/r/onbloc/qux", amount: 2000},
			},
			expectedResultAccu: `{"height":"123","now":"1234567890","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`,
		},
		{
			name: "api get accu transfer to gov staker success",
			historyItem: []struct {
				token  string
				amount uint64
			}{
				{token: "gno.land/r/onbloc/bar", amount: 1000},
				{token: "gno.land/r/onbloc/bar", amount: 2000},
			},
			expectedResultAccu: `{"height":"123","now":"1234567890","gno.land/r/onbloc/bar":"3000"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resetApi(t)

			for _, item := range tt.historyItem {
				protocolFeeState.addAccuToGovStaker(item.token, item.amount)
			}

			// when
			resultAccu := ApiGetAccuTransferToGovStaker()

			// then
			uassert.Equal(t, tt.expectedResultAccu, resultAccu)
		})
	}
}

func TestApiGetHistoryTransferToGovStaker(t *testing.T) {
	tests := []struct {
		name        string
		historyItem []struct {
			token  string
			amount uint64
		}
		expectedResultHistory string
		expectedResultAccu    string
	}{
		{
			name: "api get history transfer to gov staker success",
			historyItem: []struct {
				token  string
				amount uint64
			}{
				{token: "gno.land/r/onbloc/bar", amount: 1000},
				{token: "gno.land/r/onbloc/qux", amount: 2000},
			},
			expectedResultHistory: `{"height":"124","now":"1234567895","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`,
			expectedResultAccu:    `{"height":"124","now":"1234567895","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resetApi(t)
			testing.SkipHeights(1)

			for _, item := range tt.historyItem {
				protocolFeeState.addAccuToGovStaker(item.token, item.amount)
				protocolFeeState.updateDistributedToGovStakerHistory(item.token, item.amount)
			}

			// when
			resultHistory := ApiGetHistoryTransferToGovStaker()
			resultAccu := ApiGetAccuTransferToGovStaker()

			// then
			uassert.Equal(t, tt.expectedResultHistory, resultHistory)
			uassert.Equal(t, tt.expectedResultAccu, resultAccu)
		})
	}
}

func TestApiGetAccuTransferToDevOps(t *testing.T) {
	tests := []struct {
		name        string
		historyItem []struct {
			token  string
			amount uint64
		}
		expectedResultAccu string
	}{
		{
			name: "api get accu transfer to devOps success",
			historyItem: []struct {
				token  string
				amount uint64
			}{
				{token: "gno.land/r/onbloc/bar", amount: 1000},
				{token: "gno.land/r/onbloc/qux", amount: 2000},
			},
			expectedResultAccu: `{"height":"124","now":"1234567895","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resetApi(t)
			testing.SkipHeights(1)

			for _, item := range tt.historyItem {
				protocolFeeState.addAccuToDevOps(item.token, item.amount)
			}

			// when
			resultAccu := ApiGetAccuTransferToDevOps()

			// then
			uassert.Equal(t, tt.expectedResultAccu, resultAccu)
		})
	}
}

// This test should be run individually. It may appear as failed when running the full test suite locally,
// but there's no need to fix the test.
// In CI, the `-run` flag is used to ensure isolated execution, so there's no need to add separate skip handling.
func TestApiGetHistoryTransferToDevOps(t *testing.T) {
	tests := []struct {
		name        string
		historyItem []struct {
			token  string
			amount uint64
		}
		expectedResultHistory string
		expectedResultAccu    string
	}{
		{
			name: "api get history transfer to devOps success",
			historyItem: []struct {
				token  string
				amount uint64
			}{
				{token: "gno.land/r/onbloc/bar", amount: 1000},
				{token: "gno.land/r/onbloc/qux", amount: 2000},
			},
			expectedResultHistory: `{"height":"124","now":"1234567895","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`,
			expectedResultAccu:    `{"height":"124","now":"1234567895","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resetApi(t)
			testing.SkipHeights(1)

			for _, item := range tt.historyItem {
				protocolFeeState.addAccuToDevOps(item.token, item.amount)
				protocolFeeState.updateDistributedToDevOpsHistory(item.token, item.amount)
			}

			// when
			resultHistory := ApiGetHistoryTransferToDevOps()
			resultAccu := ApiGetAccuTransferToDevOps()

			// then
			uassert.Equal(t, tt.expectedResultHistory, resultHistory)
			uassert.Equal(t, tt.expectedResultAccu, resultAccu)
		})
	}
}

func resetApi(t *testing.T) {
	protocolFeeState = NewProtocolFeeState()
}
