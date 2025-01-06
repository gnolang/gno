package testing

import "errors"

type Seed_Queue struct {
	data []Seed
}

func New_Seed_Queue() *Seed_Queue {
	return &Seed_Queue{
		data: make([]Seed, 0),
	}
}

func (q *Seed_Queue) Enqueue(seed Seed) {
	q.data = append(q.data, seed)
}

func (q *Seed_Queue) Dequeue() (Seed, bool) {
	if len(q.data) == 0 {
		return Seed{}, false
	}

	front := q.data[0]

	q.data = q.data[1:]
	// Reduce memory usage
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]Seed, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}
	return front, true
}

func (q *Seed_Queue) Peek() (Seed, error) {
	if len(q.data) == 0 {
		return Seed{}, errors.New("queue is empty")
	}
	return q.data[0], nil
}

func (q *Seed_Queue) Select() Seed

func (q *Seed_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

func (q *Seed_Queue) Size() int {
	return len(q.data)
}

func (q *Seed_Queue) Display() {
	if len(q.data) == 0 {
		println("Queue is empty")
		return
	}
	println("Queue seeds:")
	for i, seed := range q.data {
		println("[", i, "]: {content:", seed.Content, "}")
	}
}

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

type Uint_Queue struct {
	data []uint
}

func New_Uint_Queue() *Uint_Queue {
	return &Uint_Queue{data: make([]uint, 0)}
}

func (q *Uint_Queue) Uint_Enqueue(value uint) {
	q.data = append(q.data, value)
}

// Remove from queue and return.
// If the slice is empty compared to the capacity,
// Added logic to copy into new slices to organize memory.
func (q *Uint_Queue) Uint_Dequeue() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}

	value := q.data[0]
	q.data[0] = 0

	q.data = q.data[1:]

	// Memory leak prevention
	// When the current slice length (len) is less than half the capacity (cap)
	// Copy to new slice, remove the front discarded space.
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]uint, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}

	return value, true
}

func (q *Uint_Queue) Uint_Peek() (uint, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	return q.data[0], true
}

func (q *Uint_Queue) IsEmpty() bool {
	return len(q.data) == 0
}

func (q *Uint_Queue) Size() int {
	return len(q.data)
}
