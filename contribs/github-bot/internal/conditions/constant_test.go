package conditions

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/xlab/treeprint"
)

func TestAlways(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	assert.True(t, Always().IsMet(nil, details), "condition should have a met status: true")
	assert.True(t, utils.TestLastNodeStatus(t, true, details), "condition details should have a status: true")
}

func TestNever(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	assert.False(t, Never().IsMet(nil, details), "condition should have a met status: false")
	assert.True(t, utils.TestLastNodeStatus(t, false, details), "condition details should have a status: false")
}
