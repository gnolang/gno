package fuzzinglib

// Node 구조체: 링크드 리스트의 각 노드
type Node struct {
	data uint
	next *Node
}

// Uint_LinkedList 구조체
type Uint_LinkedList struct {
	head *Node
}

// Append: 리스트 끝에 새로운 노드 추가
func (ll *Uint_LinkedList) Append(data uint) {
	newNode := &Node{data: data, next: nil}
	if ll.head == nil {
		ll.head = newNode
		return
	}

	current := ll.head
	for current.next != nil {
		current = current.next
	}
	current.next = newNode
}

// Prepend: 리스트 시작에 새로운 노드 추가
func (ll *Uint_LinkedList) Prepend(data uint) {
	newNode := &Node{data: data, next: ll.head}
	ll.head = newNode
}

// DeleteNode: 특정 노드를 직접 삭제 (O(1))
func (ll *Uint_LinkedList) DeleteNode(node *Node) {
	if node == nil || ll.head == nil {
		return
	}

	// 헤드 노드를 삭제하는 경우
	if ll.head == node {
		ll.head = ll.head.next
		return
	}

	// 이전 노드를 찾기
	current := ll.head
	for current.next != nil && current.next != node {
		current = current.next
	}

	// 노드를 삭제
	if current.next == node {
		current.next = node.next
	}
}

// Delete: 주어진 값을 가진 첫 번째 노드 삭제 (O(n))
func (ll *Uint_LinkedList) Delete(value uint) {
	if ll.head == nil {
		return
	}

	// 헤드 노드가 삭제 대상일 경우
	if ll.head.data == value {
		ll.head = ll.head.next
		return
	}

	current := ll.head
	for current.next != nil && current.next.data != value {
		current = current.next
	}

	// 값을 찾았으면 삭제
	if current.next != nil {
		current.next = current.next.next
	}
}

// Display: 리스트의 모든 노드 출력
func (ll *Uint_LinkedList) Display() {
	current := ll.head
	for current != nil {
		println("%d -> ", current.data)
		current = current.next
	}
	println("nil")
}

// SearchNode: 리스트에서 값을 검색하여 노드 반환
func (ll *Uint_LinkedList) SearchNode(value uint) *Node {
	current := ll.head
	for current != nil {
		if current.data == value {
			return current
		}
		current = current.next
	}
	return nil
}

// IsEmpty: 리스트가 비었는지 확인
func (ll *Uint_LinkedList) IsEmpty() bool {
	return ll.head == nil
}

// func main() {
// 	ll := &Uint_LinkedList{}

// 	// 노드 추가
// 	ll.Append(10)
// 	ll.Append(20)
// 	ll.Append(30)
// 	println("After appending:")
// 	ll.Display()

// 	// 리스트 앞에 노드 추가
// 	ll.Prepend(5)
// 	println("After prepending:")
// 	ll.Display()

// 	// 특정 노드 검색 후 삭제
// 	node := ll.SearchNode(20)
// 	if node != nil {
// 		println("Deleting node with value 20:")
// 		ll.DeleteNode(node) // O(1) 삭제
// 		ll.Display()
// 	} else {
// 		println("Node with value 20 not found")
// 	}

// 	// 헤드 노드 삭제
// 	node = ll.SearchNode(5)
// 	if node != nil {
// 		println("Deleting head node with value 5:")
// 		ll.DeleteNode(node) // O(1) 삭제
// 		ll.Display()
// 	}

// 	// 없는 값 삭제 시도
// 	node = ll.SearchNode(40)
// 	if node != nil {
// 		ll.DeleteNode(node)
// 	} else {
// 		println("Node with value 40 not found")
// 	}
// }
