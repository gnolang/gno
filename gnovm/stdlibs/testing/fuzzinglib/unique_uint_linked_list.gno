package fuzzinglib

//! data는 유니크해야 함!!!
//! 그걸 이용해서 시간복잡도 줄인 리스트기 때문!!

// Node: 단일 연결 리스트의 노드
type Node struct {
	data uint
	next *Node
}

// Unique_Uint_LinkedList: Singly Linked List + O(1) 삽입/삭제를 위한 자료구조
type Unique_Uint_LinkedList struct {
	head      *Node
	tail      *Node
	nodeMap   map[uint]*Node  // data -> node
	parentMap map[*Node]*Node // node -> parent node
}

// New_Uint_LinkedList: 생성자
func New_Unique_Uint_LinkedList() *Unique_Uint_LinkedList {
	return &Unique_Uint_LinkedList{
		head:      nil,
		tail:      nil,
		nodeMap:   make(map[uint]*Node),
		parentMap: make(map[*Node]*Node),
	}
}

// -------------------------------------------------------
// 1) Append: 리스트 끝에 새 노드를 O(1)에 추가
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

	// 이미 노드가 있는 경우 (tail 존재)
	ll.tail.next = newNode
	ll.parentMap[newNode] = ll.tail // 새 노드의 parent는 기존 tail
	ll.tail = newNode               // tail 업데이트
	ll.nodeMap[data] = newNode
}

// -------------------------------------------------------
// 2) Prepend: 리스트 머리에 새 노드를 O(1)에 추가
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) Prepend(data uint) {
	newNode := &Node{data: data}

	// 1) 빈 리스트였다면 (head == nil)
	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
		ll.nodeMap[data] = newNode
		ll.parentMap[newNode] = nil
		return
	}

	// 2) 빈 리스트가 아니라면
	newNode.next = ll.head
	// 기존 head(옛 머리 노드)의 parent를 newNode로 갱신
	ll.parentMap[ll.head] = newNode

	// 새 머리 노드로 교체
	ll.head = newNode

	// 맵 갱신
	ll.nodeMap[data] = newNode
	ll.parentMap[newNode] = nil
}

// -------------------------------------------------------
// 3) Delete: data == value 인 노드를 O(1)에 삭제
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) Delete(value uint) {
	targetNode, ok := ll.nodeMap[value]
	if !ok {
		// 해당 value를 가진 노드가 없으면 그냥 리턴
		return
	}

	// (1) 삭제 노드가 head인 경우
	if targetNode == ll.head {
		ll.head = ll.head.next
		// 만약 head가 nil이 됐다면 리스트가 비었다는 것
		if ll.head == nil {
			ll.tail = nil
		} else {
			ll.parentMap[ll.head] = nil
		}
		delete(ll.nodeMap, value)
		delete(ll.parentMap, targetNode)
		return
	}

	// (2) head가 아닌 일반 노드인 경우
	parent := ll.parentMap[targetNode]
	if parent == nil {
		// 논리적으로는 발생하지 않음(오류 처리)
		return
	}
	// parent -> targetNode -> next 연결 끊기
	parent.next = targetNode.next

	// 만약 삭제하는 노드가 tail이었다면 tail 갱신
	if targetNode == ll.tail {
		ll.tail = parent
	} else {
		// tail이 아니면 targetNode.next가 존재할 수 있으므로 parentMap 갱신
		ll.parentMap[targetNode.next] = parent
	}

	// nodeMap, parentMap 정리
	delete(ll.nodeMap, value)
	delete(ll.parentMap, targetNode)
}

// -------------------------------------------------------
// 4) DeleteNode: 노드 포인터로 직접 O(1)에 삭제
// -------------------------------------------------------
func (ll *Unique_Uint_LinkedList) DeleteNode(node *Node) {
	if node == nil {
		return
	}

	// (1) node가 head인 경우
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

	// (2) head가 아닌 경우
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

// -------------------------------------------------------
// 5) 기타 메서드들
// -------------------------------------------------------

// SearchNode: data -> nodeMap으로 O(1) 검색
func (ll *Unique_Uint_LinkedList) SearchNode(value uint) *Node {
	return ll.nodeMap[value]
}

// Display: 리스트의 모든 노드 data를 순서대로 출력
func (ll *Unique_Uint_LinkedList) Display() {
	current := ll.head
	for current != nil {
		println("%d -> ", current.data)
		current = current.next
	}
	println("nil")
}

// IsEmpty: 리스트가 비었는지 확인
func (ll *Unique_Uint_LinkedList) IsEmpty() bool {
	return ll.head == nil
}

func (ll *Unique_Uint_LinkedList) Peek() (uint, bool) {
	if ll.head == nil {
		return 0, false
	}
	return ll.head.data, true
}

/*
// -------------------------------------------------------
// 간단한 테스트 예시
// -------------------------------------------------------
func main() {
    ll := New_Unique_Uint_LinkedList()

    // 노드 추가 (Append: O(1))
    ll.Append(10)
    ll.Append(20)
    ll.Append(30)
    println("After appending:")
    ll.Display()

    // 리스트 앞에 노드 추가 (Prepend: O(1))
    ll.Prepend(5)
    println("After prepending:")
    ll.Display()

    // Delete(20) => O(1) 삭제
    println("Deleting node with value 20:")
    ll.Delete(20)
    ll.Display()

    // head(5) 삭제 => O(1)
    println("Deleting head node with value 5:")
    ll.Delete(5)
    ll.Display()

    // 없는 값(40) 삭제 시도
    println("Deleting node with value 40 (not exist):")
    ll.Delete(40)
    ll.Display()
}
*/
