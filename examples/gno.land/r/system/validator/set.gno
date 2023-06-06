package validator

type Set struct {
	ActiveNodes avl.Tree // owner -> Node
	// InactiveNodes avl.Tree // owner -> Node
}
