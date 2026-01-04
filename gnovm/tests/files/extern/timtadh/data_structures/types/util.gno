package types

// XXX modified by gno
func IsNil(object TreeNode) bool {
	return object == nil || object.IsNil()
}

func MakeKVIteratorFromTreeNodeIterator(tni TreeNodeIterator) KVIterator {
	var kv_iterator KVIterator
	kv_iterator = func() (key Hashable, value any, next KVIterator) {
		var tn TreeNode
		tn, tni = tni()
		if tni == nil {
			return nil, nil, nil
		}
		return tn.Key(), tn.Value(), kv_iterator
	}
	return kv_iterator
}

func ChainTreeNodeIterators(tnis ...TreeNodeIterator) TreeNodeIterator {
	var make_tni func(int) TreeNodeIterator
	make_tni = func(i int) (tni_iterator TreeNodeIterator) {
		if i >= len(tnis) {
			return nil
		}
		var next TreeNodeIterator = tnis[i]
		tni_iterator = func() (TreeNode, TreeNodeIterator) {
			var tn TreeNode
			tn, next = next()
			if next == nil {
				tni := make_tni(i + 1)
				if tni == nil {
					return nil, nil
				} else {
					return tni()
				}
			}
			return tn, tni_iterator
		}
		return tni_iterator
	}
	return make_tni(0)
}

func MakeKeysIterator(obj KVIterable) KIterator {
	kv_iterator := obj.Iterate()
	var k_iterator KIterator
	k_iterator = func() (key Hashable, next KIterator) {
		key, _, kv_iterator = kv_iterator()
		if kv_iterator == nil {
			return nil, nil
		}
		return key, k_iterator
	}
	return k_iterator
}

func MakeValuesIterator(obj KVIterable) Iterator {
	kv_iterator := obj.Iterate()
	var v_iterator Iterator
	v_iterator = func() (value any, next Iterator) {
		_, value, kv_iterator = kv_iterator()
		if kv_iterator == nil {
			return nil, nil
		}
		return value, v_iterator
	}
	return v_iterator
}

func MakeItemsIterator(obj KVIterable) (kit KIterator) {
	kv_iterator := obj.Iterate()
	kit = func() (item Hashable, next KIterator) {
		var key Hashable
		var value any
		key, value, kv_iterator = kv_iterator()
		if kv_iterator == nil {
			return nil, nil
		}
		return &MapEntry{key, value}, kit
	}
	return kit
}

func make_child_slice(node BinaryTreeNode) []BinaryTreeNode {
	nodes := make([]BinaryTreeNode, 0, 2)
	if !IsNil(node) {
		if !IsNil(node.Left()) {
			nodes = append(nodes, node.Left())
		}
		if !IsNil(node.Right()) {
			nodes = append(nodes, node.Right())
		}
	}
	return nodes
}

func DoGetChild(node BinaryTreeNode, i int) TreeNode {
	return make_child_slice(node)[i]
}

func DoChildCount(node BinaryTreeNode) int {
	return len(make_child_slice(node))
}

func MakeChildrenIterator(node BinaryTreeNode) TreeNodeIterator {
	nodes := make_child_slice(node)
	var make_tn_iterator func(int) TreeNodeIterator
	make_tn_iterator = func(i int) TreeNodeIterator {
		return func() (kid TreeNode, next TreeNodeIterator) {
			if i < len(nodes) {
				return nodes[i], make_tn_iterator(i + 1)
			}
			return nil, nil
		}
	}
	return make_tn_iterator(0)
}
