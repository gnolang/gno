package condition

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/utils"
	"github.com/xlab/treeprint"
)

func TestAlways(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	if !Always().IsMet(nil, details) {
		t.Errorf("condition should have a met status: %t", true)
	}
	if !utils.TestLastNodeStatus(t, true, details) {
		t.Errorf("condition details should have a status: %t", true)
	}
}

func TestNever(t *testing.T) {
	t.Parallel()

	details := treeprint.New()
	if Never().IsMet(nil, details) {
		t.Errorf("condition should have a met status: %t", false)
	}
	if !utils.TestLastNodeStatus(t, false, details) {
		t.Errorf("condition details should have a status: %t", false)
	}
}
