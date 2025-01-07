package testing

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type supportedType string

const (
	ByteArray supportedType = "[]byte"
	String    supportedType = "string"
	Bool      supportedType = "bool"
	Byte      supportedType = "byte"
	Rune      supportedType = "rune"
	Float32   supportedType = "float32"
	Float64   supportedType = "float64"
	Int       supportedType = "int"
	Int8      supportedType = "int8"
	Int16     supportedType = "int16"
	Int32     supportedType = "int32"
	Int64     supportedType = "int64"
	Uint      supportedType = "uint"
	Uint8     supportedType = "uint8"
	Uint16    supportedType = "uint16"
	Uint32    supportedType = "uint32"
	Uint64    supportedType = "uint64"
)

type seed struct {
	pid           uint
	id            uint
	gen           uint
	isCoordinated bool
	hashNumber    HashNumber
	content       []interface{}
	result        interface{}
}

const (
	// Effective when path kinds(coverage) is in range 300,000~1,000,000.
	// and effiective when iters is less than 100_000_000
	// More than that is possible, but less effective.
	MaxCapacity     uint64 = 100_000
	PartialCapacity        = MaxCapacity / 5
)

type stateMachine struct {
	trials     uint
	inputCount uint

	allCoverage     Coverage // []coveredline
	coveredCoverage Coverage // []coveredline

	// for f.Add in f.Fuzz
	seedType             []supportedType
	inithashnumber       []HashNumber
	stringByteCandidates []int

	// for f.Fuzz
	priorityCache          uint
	hashNumber2Seeds       *[MaxCapacity]*seedQueue
	hashNumber2Priority    *[MaxCapacity]uint
	priority2HashNumberULL *[MaxCapacity]*uniquesLinkedList

	crashLogger           *crashLogger
	verbose               bool
	inspectingHashNumbers uint
}

func newStateMachine(trials uint, inputCount uint) *stateMachine {
	settedAllULL := func() *[MaxCapacity]*uniquesLinkedList {
		arr := &[MaxCapacity]*uniquesLinkedList{}
		for i := range arr {
			arr[i] = newUniquesLinkedList()
		}
		return arr
	}()
	return &stateMachine{
		inputCount:      inputCount,
		trials:          trials,
		allCoverage:     Coverage{{coName: "nil", coLine: 0}},
		coveredCoverage: Coverage{{coName: "nil", coLine: 0}},

		priorityCache: 1,

		hashNumber2Priority:    &[MaxCapacity]uint{},
		hashNumber2Seeds:       &[MaxCapacity]*seedQueue{},
		priority2HashNumberULL: settedAllULL,

		inspectingHashNumbers: 0,
		crashLogger:           newCrashLogger(),
	}
}

type endInfo struct {
	completeTrials bool
	maxedCAPACITY  bool
}

func (sm *stateMachine) coordinateAbstraction(hn HashNumber, abstractNumber uint) endInfo {
	sm.inputCount = sm.inputCount + abstractNumber
	// Get current priorty
	// 1. Prior increases discontinuously
	// 2. Hn must increase continuously
	sm.hashNumber2Priority[hn] = sm.hashNumber2Priority[hn] + abstractNumber
	nextPriority := sm.hashNumber2Priority[hn]
	nextHashNumber := hn
	if sm.verbose {
		if (sm.inputCount-abstractNumber)%2000 > sm.inputCount%2000 {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	} else {
		if (sm.inputCount-abstractNumber)%(sm.trials/5) > sm.inputCount%(sm.trials/5) {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	}
	if sm.inputCount >= sm.trials {
		return endInfo{
			completeTrials: true,
			maxedCAPACITY:  false,
		}
	} else if uint64(nextPriority) >= (MaxCapacity-1) || uint64(nextHashNumber) >= (MaxCapacity-1) {
		// Case 1: Priority reaches MAX (and protect overflow)
		// Case 2 HashNumber reaches Max (and protect overflow)
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  true,
		}
	} else {
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
}

func (sm *stateMachine) coordinateSeed(seed seed) endInfo {
	if seed.isCoordinated {
		sm.hashNumber2Seeds[seed.hashNumber].enqueue(seed)
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
	hn := seed.hashNumber
	seed.isCoordinated = true
	sm.inputCount++

	if sm.hashNumber2Seeds[hn] == nil {
		sm.inspectingHashNumbers++
		sm.hashNumber2Seeds[hn] = newSeedQueue()
	}
	sm.hashNumber2Seeds[hn].enqueue(seed)
	oldPriority := sm.hashNumber2Priority[hn]
	sm.hashNumber2Priority[hn]++
	updatedPriority := sm.hashNumber2Priority[hn]

	if updatedPriority == 1 {
		sm.priorityCache = 1
		sm.priority2HashNumberULL[updatedPriority].append(uint(hn))
	} else {
		sm.priority2HashNumberULL[oldPriority].delete(uint(hn))
		sm.priority2HashNumberULL[updatedPriority].append(uint(hn))
	}
	if sm.verbose {
		if sm.inputCount%2000 == 0 {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	} else {
		if sm.inputCount%(sm.trials/5) == 0 {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	}

	if sm.inputCount >= sm.trials {
		return endInfo{
			completeTrials: true,
			maxedCAPACITY:  false,
		}
	} else if uint64(updatedPriority) >= (MaxCapacity-1) || uint64(hn) >= (MaxCapacity-1) {
		// Case 1: Priority reaches MAX
		// Case 2 HashNumber reaches Max
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  true,
		}
	} else {
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
}

func (sm *stateMachine) popInitSeedByHN(hn HashNumber) seed {
	popedSeed, isOnce := sm.hashNumber2Seeds[hn].dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	return popedSeed
}

func (sm *stateMachine) popSeeds() []seed {
	var highestProrityHashNumber uint

	for {

		hn, isExist := sm.priority2HashNumberULL[sm.priorityCache].peek()
		if !isExist {
			sm.priorityCache++

			continue
		}

		highestProrityHashNumber = hn
		break
	}

	popedSeed1, isOnce := sm.hashNumber2Seeds[highestProrityHashNumber].dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	peekedSeed2, err := sm.hashNumber2Seeds[highestProrityHashNumber].peek()
	if err != nil {
		return []seed{popedSeed1}
	}
	pid1 := popedSeed1.pid
	pid2 := peekedSeed2.pid
	if pid1 == pid2 {
		popedSeed2, _ := sm.hashNumber2Seeds[highestProrityHashNumber].dequeue()
		return []seed{popedSeed1, popedSeed2}
	} else {
		return []seed{popedSeed1}
	}
}

type priorytyAndHashNumber struct {
	priority   uint
	hashNumber int
}
type priorytyHashNumberList []priorytyAndHashNumber

func (phl priorytyHashNumberList) Len() int           { return len(phl) }
func (phl priorytyHashNumberList) Less(i, j int) bool { return phl[i].priority < phl[j].priority }
func (phl priorytyHashNumberList) Swap(i, j int)      { phl[i], phl[j] = phl[j], phl[i] }

type seedAndPrioryty struct {
	seed     seed
	priority uint
}

func (sm *stateMachine) summarize() []seed {
	collectedPriorytyAndHN := priorytyHashNumberList{}
	for i, e := range sm.hashNumber2Priority {

		if e == 0 {
			break
		}
		collectedPriorytyAndHN = append(collectedPriorytyAndHN, priorytyAndHashNumber{
			priority:   e,
			hashNumber: i,
		})
		// Exit when slice length exceeds Partial_Capacity
		if len(collectedPriorytyAndHN) >= int(PartialCapacity) {
			break
		}
	}

	sort.Sort(collectedPriorytyAndHN)

	sampledSeedAndPrioryty := []seedAndPrioryty{}

	for _, pair := range collectedPriorytyAndHN {
		hn := pair.hashNumber
		seed := sm.hashNumber2Seeds[hn].data[0]
		priority := pair.priority
		sampledSeedAndPrioryty = append(sampledSeedAndPrioryty, seedAndPrioryty{seed, priority})
	}

	total := len(sampledSeedAndPrioryty)

	summarizedSeeds := []seed{}

	for i, seedAndPriority := range sampledSeedAndPrioryty {

		rank := i + 1

		seed := seedAndPriority.seed
		seed.pid = 0
		seed.id = uint(i)
		seed.gen = allocateGen(seedAndPriority.priority, rank, total)
		seed.isCoordinated = false

		summarizedSeeds = append(summarizedSeeds, seed)
	}
	return summarizedSeeds
}

func allocateGen(priority uint, rank int, total int) uint {
	var bigger uint

	switch rank {
	case 1:
		bigger = 1
	case 2:
		bigger = 2
	case 3:
		bigger = 3
	case 4:
		bigger = 4
	case 5:
		bigger = 5
	default:
		bigger = 15
	}
	c1CutLine := uint(float64(MaxCapacity) * 0.008)
	c2CutLine := uint(float64(MaxCapacity) * 0.014)
	c3CutLine := uint(float64(MaxCapacity) * 0.02)
	c4CutLine := uint(float64(MaxCapacity) * 0.03)
	c5CutLine := uint(float64(MaxCapacity) * 0.08)
	c6CutLine := uint(float64(MaxCapacity) * 0.1)
	c7CutLine := uint(float64(MaxCapacity) * 0.15)
	c8CutLine := uint(float64(MaxCapacity) * 0.30)
	switch {
	case priority <= c1CutLine:
		if bigger > 1 {
			bigger = 1
		}
	case priority <= c2CutLine:
		if bigger > 2 {
			bigger = 2
		}
	case priority <= c3CutLine:
		if bigger > 3 {
			bigger = 3
		}
	case priority <= c4CutLine:
		if bigger > 4 {
			bigger = 4
		}
	case priority <= c5CutLine:
		if bigger > 5 {
			bigger = 5
		}
	case priority <= c6CutLine:
		if bigger > 6 {
			bigger = 6
		}
	case priority <= c7CutLine:
		if bigger > 7 {
			bigger = 7
		}
	case priority <= c8CutLine:
		if bigger > 10 {
			bigger = 10
		}
	default:

		if bigger > 15 {
			bigger = 15
		}
	}
	return bigger
}

type testResult struct {
	panicOccurred bool
	panicMessage  string
	prror         error
}

type crashCase struct {
	hashNumber   uint
	input        string
	rawContent   []interface{}
	inputCount   uint
	isPanic      bool
	panicMessage string
	errorMsg     string
	timestamp    time.Time
}

// crash logger
type crashLogger struct {
	data []crashCase
}

func newCrashLogger() *crashLogger {
	return &crashLogger{
		data: make([]crashCase, 0),
	}
}

func (c *crashLogger) addCase(seed seed, r testResult) {
	var crashcase crashCase
	if r.prror == nil {
		crashcase = crashCase{
			hashNumber:   uint(seed.hashNumber),
			input:        contentToString(seed.content),
			rawContent:   seed.content,
			inputCount:   seed.id,
			isPanic:      r.panicOccurred,
			panicMessage: r.panicMessage,
			errorMsg:     "",
			timestamp:    time.Now(),
		}
	} else {
		crashcase = crashCase{
			hashNumber:   uint(seed.hashNumber),
			input:        contentToString(seed.content),
			rawContent:   seed.content,
			inputCount:   seed.id,
			isPanic:      r.panicOccurred,
			panicMessage: r.panicMessage,
			errorMsg:     strings.TrimSpace(r.prror.Error()),
			timestamp:    time.Now(),
		}
	}
	c.data = append(c.data, crashcase)
}

// data structures
type seedQueue struct {
	data []seed
}

func newSeedQueue() *seedQueue {
	return &seedQueue{
		data: make([]seed, 0),
	}
}

func (q *seedQueue) enqueue(seed seed) {
	q.data = append(q.data, seed)
}

func (q *seedQueue) dequeue() (seed, bool) {
	if len(q.data) == 0 {
		return seed{}, false
	}

	front := q.data[0]

	q.data = q.data[1:]
	// Reduce memory usage
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]seed, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}
	return front, true
}

func (q *seedQueue) peek() (seed, error) {
	if len(q.data) == 0 {
		return seed{}, errors.New("queue is empty")
	}
	return q.data[0], nil
}

// ! Data must be unique!!!
// ! It's a list that reduces time complexity by using it!!
// node: node in single connection list
type node struct {
	data uint
	next *node
}

// uniquesLinkedList: Data Structure for Singly Linked List + O(1) Insertion/Deletion
type uniquesLinkedList struct {
	head      *node
	tail      *node
	nodeMap   map[uint]*node  // data -> node
	parentMap map[*node]*node // node -> parent node
}

func newUniquesLinkedList() *uniquesLinkedList {
	return &uniquesLinkedList{
		head:      nil,
		tail:      nil,
		nodeMap:   make(map[uint]*node),
		parentMap: make(map[*node]*node),
	}
}

// -------------------------------------------------------
// 1) Add a new node to O(1) at the end of the Appendix list
// -------------------------------------------------------
func (ll *uniquesLinkedList) append(data uint) {
	newNode := &node{data: data, next: nil}

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
func (ll *uniquesLinkedList) prepend(data uint) {
	newNode := &node{data: data}

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
func (ll *uniquesLinkedList) delete(value uint) {
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
func (ll *uniquesLinkedList) deleteNode(node *node) {
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

func (ll *uniquesLinkedList) searchNode(value uint) *node {
	return ll.nodeMap[value]
}

func (ll *uniquesLinkedList) display() {
	current := ll.head
	for current != nil {
		println("%d -> ", current.data)
		current = current.next
	}
	println("nil")
}

func (ll *uniquesLinkedList) isEmpty() bool {
	return ll.head == nil
}

func (ll *uniquesLinkedList) peek() (uint, bool) {
	if ll.head == nil {
		return 0, false
	}
	return ll.head.data, true
}
