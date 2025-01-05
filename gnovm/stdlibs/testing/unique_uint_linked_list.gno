package testing

// ! Data must be unique!!!
// ! It's a list that reduces time complexity by using it!!
// Node: Node in single connection list
type Node struct {
	data uint
	next *Node
}

// Unique_Uint_LinkedList: Data Structure for Singly Linked List + O(1) Insertion/Deletion
type Unique_Uint_LinkedList struct {
	head      *Node
	tail      *Node
	nodeMap   map[uint]*Node  // data -> node
	parentMap map[*Node]*Node // node -> parent node
}

func New_Unique_Uint_LinkedList() *Unique_Uint_LinkedList {
	return &Unique_Uint_LinkedList{
		head:      nil,
		tail:      nil,
		nodeMap:   make(map[uint]*Node),
		parentMap: make(map[*Node]*Node),
	}
}

// -------------------------------------------------------
// 1) Add a new node to O(1) at the end of the Appendix list
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) Append(data uint) {
	newNode := &Node{data: data, next: nil}

	// 빈 리스트인 경우
	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
		ll.nodeMap[data] = newNode
		ll.parentMap[newNode] = nil
		return
	}

	ll.tail.next = newNode
	ll.parentMap[newNode] = ll.tail
	ll.tail = newNode
	ll.nodeMap[data] = newNode
}

// -------------------------------------------------------
// 2) Prepend: Add a new node to O(1) in the head of the list
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) Prepend(data uint) {
	newNode := &Node{data: data}

	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
		ll.nodeMap[data] = newNode
		ll.parentMap[newNode] = nil
		return
	}

	newNode.next = ll.head

	ll.parentMap[ll.head] = newNode

	ll.head = newNode

	ll.nodeMap[data] = newNode
	ll.parentMap[newNode] = nil
}

// -------------------------------------------------------
// 3) Delete: Delete node with data == value to O(1)
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) Delete(value uint) {
	targetNode, ok := ll.nodeMap[value]
	if !ok {
		return
	}

	if targetNode == ll.head {
		ll.head = ll.head.next
		if ll.head == nil {
			ll.tail = nil
		} else {
			ll.parentMap[ll.head] = nil
		}
		delete(ll.nodeMap, value)
		delete(ll.parentMap, targetNode)
		return
	}

	parent := ll.parentMap[targetNode]
	if parent == nil {
		return
	}
	parent.next = targetNode.next

	if targetNode == ll.tail {
		ll.tail = parent
	} else {
		ll.parentMap[targetNode.next] = parent
	}

	delete(ll.nodeMap, value)
	delete(ll.parentMap, targetNode)
}

// -------------------------------------------------------
// 4) DeleteNode: Delete directly to O(1) with node pointer
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) DeleteNode(node *Node) {
	if node == nil {
		return
	}

	if node == ll.head {
		ll.head = ll.head.next
		if ll.head == nil {
			ll.tail = nil
		} else {
			ll.parentMap[ll.head] = nil
		}
		delete(ll.nodeMap, node.data)
		delete(ll.parentMap, node)
		return
	}

	parent := ll.parentMap[node]
	if parent == nil {
		return
	}
	parent.next = node.next

	if node == ll.tail {
		ll.tail = parent
	} else {
		ll.parentMap[node.next] = parent
	}

	delete(ll.nodeMap, node.data)
	delete(ll.parentMap, node)
}

func (ll *Unique_Uint_LinkedList) SearchNode(value uint) *Node {
	return ll.nodeMap[value]
}

func (ll *Unique_Uint_LinkedList) Display() {
	current := ll.head
	for current != nil {
		println("%d -> ", current.data)
		current = current.next
	}
	println("nil")
}

func (ll *Unique_Uint_LinkedList) IsEmpty() bool {
	return ll.head == nil
}

func (ll *Unique_Uint_LinkedList) Peek() (uint, bool) {
	if ll.head == nil {
		return 0, false
	}
	return ll.head.data, true
}
