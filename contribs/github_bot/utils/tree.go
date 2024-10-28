package utils

import (
	"fmt"

	"github.com/xlab/treeprint"
)

func AddStatusNode(b bool, desc string, details treeprint.Tree) bool {
	if b {
		details.AddNode(fmt.Sprintf("🟢 %s", desc))
	} else {
		details.AddNode(fmt.Sprintf("🔴 %s", desc))
	}

	return b
}
