package utils

import (
	"fmt"

	"github.com/xlab/treeprint"
)

type Status string

const (
	Success Status = "🟢"
	Fail    Status = "🔴"
)

func AddStatusNode(b bool, desc string, details treeprint.Tree) bool {
	if b {
		details.AddNode(fmt.Sprintf("%s %s", Success, desc))
	} else {
		details.AddNode(fmt.Sprintf("%s %s", Fail, desc))
	}

	return b
}
