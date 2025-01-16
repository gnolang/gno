package utils

import (
	"strings"
	"testing"

	"github.com/xlab/treeprint"
)

func TestLastNodeStatus(t *testing.T, success bool, details treeprint.Tree) bool {
	t.Helper()

	detail := details.FindLastNode().(*treeprint.Node).Value.(string)
	status := Fail

	if success {
		status = Success
	}

	return strings.HasPrefix(detail, string(status))
}
