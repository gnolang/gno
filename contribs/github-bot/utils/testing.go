package utils

import (
	"strings"
	"testing"

	"github.com/xlab/treeprint"
)

func TestLastNodeStatus(t *testing.T, success bool, details treeprint.Tree) bool {
	t.Helper()

	detail := details.FindLastNode().(*treeprint.Node).Value.(string)

	if success {
		return strings.HasPrefix(detail, StatusSuccess)
	}
	return strings.HasPrefix(detail, StatusFail)
}
