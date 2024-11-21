package utils

import (
	"fmt"

	"github.com/xlab/treeprint"
)

const (
	StatusSuccess = "🟢"
	StatusFail    = "🔴"
)

func AddStatusNode(b bool, desc string, details treeprint.Tree) bool {
	if b {
		details.AddNode(fmt.Sprintf("%s %s", StatusSuccess, desc))
	} else {
		details.AddNode(fmt.Sprintf("%s %s", StatusFail, desc))
	}

	return b
}
