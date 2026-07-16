package iavl

import (
	"fmt"
	"strings"
)

// PrintTree prints the whole tree in an indented form.
func PrintTree(tree *ImmutableTree) {
	ndb, root := tree.ndb, tree.root
	printNode(ndb, root, 0) //nolint:errcheck
}

func printNode(ndb *nodeDB, node *Node, indent int) error {
	var indentPrefix strings.Builder
	for range indent {
		indentPrefix.WriteString("    ")
	}

	if node == nil {
		fmt.Printf("%s<nil>\n", indentPrefix.String())
		return nil
	}
	if node.rightNode != nil {
		printNode(ndb, node.rightNode, indent+1) //nolint:errcheck
	} else if node.rightNodeKey != nil {
		rightNode, err := ndb.GetNode(node.rightNodeKey)
		if err != nil {
			return err
		}
		printNode(ndb, rightNode, indent+1) //nolint:errcheck
	}

	hash := node._hash(node.nodeKey.version)

	fmt.Printf("%sh:%X\n", indentPrefix.String(), hash)
	if node.isLeaf() {
		fmt.Printf("%s%X:%X (%v)\n", indentPrefix.String(), node.key, node.value, node.subtreeHeight)
	}

	if node.leftNode != nil {
		err := printNode(ndb, node.leftNode, indent+1)
		if err != nil {
			return err
		}
	} else if node.leftNodeKey != nil {
		leftNode, err := ndb.GetNode(node.leftNodeKey)
		if err != nil {
			return err
		}
		err = printNode(ndb, leftNode, indent+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func maxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
