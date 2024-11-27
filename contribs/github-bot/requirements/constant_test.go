package requirements

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/stretchr/testify/assert"
	"github.com/xlab/treeprint"
)

func TestAlways(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	assert.True(t, Always().IsSatisfied(nil, details), "requirement should have a satisfied status: true")
	assert.True(t, utils.TestLastNodeStatus(t, true, details), "requirement details should have a status: true")
}

func TestNever(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	assert.False(t, Never().IsSatisfied(nil, details), "requirement should have a satisfied status: false")
	assert.True(t, utils.TestLastNodeStatus(t, false, details), "requirement details should have a status: false")
}
