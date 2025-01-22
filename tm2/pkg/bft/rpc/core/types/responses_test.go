package core_types

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
)

func TestStatusIndexer(t *testing.T) {
	t.Parallel()

	var status *ResultStatus
	assert.False(t, status.TxIndexEnabled())

	status = &ResultStatus{}
	assert.False(t, status.TxIndexEnabled())

	status.NodeInfo = types.NodeInfo{}
	assert.False(t, status.TxIndexEnabled())

	cases := []struct {
		expected bool
		other    types.NodeInfoOther
	}{
		{false, types.NodeInfoOther{}},
		{false, types.NodeInfoOther{TxIndex: "aa"}},
		{false, types.NodeInfoOther{TxIndex: "off"}},
		{true, types.NodeInfoOther{TxIndex: "on"}},
	}

	for _, tc := range cases {
		status.NodeInfo.Other = tc.other
		assert.Equal(t, tc.expected, status.TxIndexEnabled())
	}
}
