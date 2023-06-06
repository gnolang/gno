package validator

type Set struct {
	activeByOwner avl.Tree // ownerAddr -> *Node
	activeByAddr  avl.Tree // nodeAddr -> *Node
	// keep track of inactive?
}
