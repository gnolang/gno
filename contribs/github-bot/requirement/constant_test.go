package requirement

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/xlab/treeprint"
)

func TestAlways(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	if !Always().IsSatisfied(nil, details) {
		t.Errorf("requirement should have a satisfied status: %t", true)
	}
	if !utils.TestLastNodeStatus(t, true, details) {
		t.Errorf("requirement details should have a status: %t", true)
	}
}

func TestNever(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	if Never().IsSatisfied(nil, details) {
		t.Errorf("requirement should have a satisfied status: %t", false)
	}
	if !utils.TestLastNodeStatus(t, false, details) {
		t.Errorf("requirement details should have a status: %t", false)
	}
}
