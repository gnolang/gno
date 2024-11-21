package utils

import (
	"strings"
	"testing"

	"github.com/xlab/treeprint"
)

func TestLastNodeStatus(t *testing.T, success bool, details treeprint.Tree) bool {
	t.Helper()

	detail := details.FindLastNode().(*treeprint.Node).Value.(string)
	status := StatusFail

	if success {
		status = StatusSuccess
	}

	return strings.HasPrefix(detail, status)
}
