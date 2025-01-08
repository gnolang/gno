package testing

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	// Effective when path kinds(coverage) is in range 300,000~1,000,000.
	// and effiective when iters is less than 100_000_000
	// More than that is possible, but less effective.
	// 퍼즈매니져가 내부적으로 관리하는 배열들의 길이입니다.
	// 시드 간의 정렬,out,in을 o(1)시간 내로 다루기 위해서 배열의 최대 길이를 제한했습니다.(그러나 입력은 무한대로 받을 수 있습니다)
	MaxCapacity     uint64 = 100_000
	PartialCapacity        = MaxCapacity / 5
)

// 퍼져가 '다음에 입력에 쓸 인풋값'을 빠르게 찾아내기 위해서 입력된 시드들을 구조적으로 정리하는 매니져입니다.
// 내부 배열을 이용하여 새 입력값이 추가될 때마다 이들을 우선순위에 맞게 정렬하고, 다음 입력이 필요할 때 높은 우선순위의 입력을 제공합니다.
type fuzzManager struct {
	trials     uint // 퍼즈가 실패할 때까지 시도할 수 있는 최대 입력 수입니다. cli를 통해 정해집니다.
	inputCount uint // 현재까지 시도한 입력의 수입니다.

	seedType       []supportedType // 시드의 타입
	inithashnumber []hashNumber    // f.Add때 들어온 "진짜 시드"들을 탐색하기 위한 넘버입니다.

	// 시드를 배치, 탐색하기 위한 배열들
	priorityCache         uint                        // 최근 가장 0과 가까웠던 우선순위 값
	hashNumber2Seeds      *[MaxCapacity]*seedQueue    // 같은 해시넘버를 지니는 시드들을 모아둡니다.
	hashNumber2Priority   *[MaxCapacity]uint          // 해시넘버의 우선순위를 기록합니다
	priority2HashNumSet   *[MaxCapacity]hashNumberSet // 같은 우선순위를 지니는 해시넘버의 집합
	verbose               bool                        // cli에서 받은 verbose값. 로깅의 빈도를 정합니다.
	inspectingHashNumbers uint                        // 현재 매니져가 관리중인 해시넘버의 수수
}

// 시드를 퍼즈매니져의 내부 배열에 배치합니다.
// 배치된 시드는 '다음 입력'때 선발될 수 있습니다.
// '이 시드와 유사한 시드'가 현재까지 얼마나 입력되었었는지에 기반하여 해당 시드의 우선순위를 정합니다.
// 우선순위가 낮은, 그러니까 priority가 큰 시드는 다음 번 선택에서 선택될 확률이 줄어듭니다.
func (fm *fuzzManager) coordinateSeed(seed seed) endInfo {
	if seed.isCoordinated {
		fm.hashNumber2Seeds[seed.hashNumber].enqueue(seed)
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
	hn := seed.hashNumber
	seed.isCoordinated = true
	fm.inputCount++

	// OS에서 쓰는 aging기법과 유사합니다. 시드의 해시넘버를 받아, 그 해시넘버의 우선순위를 떨어뜨립니다.
	// 그렇게 함으로써 많이 입력된 값 집합일수록 다음번에는 더 적게 선택되게 됩니다.
	oldPriority := fm.hashNumber2Priority[hn]
	fm.hashNumber2Priority[hn]++
	updatedPriority := fm.hashNumber2Priority[hn]

	if oldPriority == 0 {
		fm.inspectingHashNumbers++
		fm.priorityCache = updatedPriority
		fm.hashNumber2Seeds[hn] = newSeedQueue()
	} else {
		fm.priority2HashNumSet[oldPriority].remove(hn)
	}
	fm.hashNumber2Seeds[hn].enqueue(seed)
	if fm.priority2HashNumSet[updatedPriority] == nil {
		fm.priority2HashNumSet[updatedPriority] = newHashNumberSet()
	}
	fm.priority2HashNumSet[updatedPriority].add(hn)
	fm.conditionallyLogProgress(1)
	return fm.evaluateEndState(updatedPriority, hn)
}

// 퍼즈는 시간이 오래걸리는 프로세스입니다. 그렇기에, 프로세스 중간중간 그 작업의 진행도를 알려줄 필요가 있습니다.
// 입력 개수를 기반으로 하여 입력 개수가 일정 수를 초과할 때마다 진행상황을 프린트합니다.
// cli에서 받는 verbose에 따라 그 빈도가 달라집니다.
func (fm *fuzzManager) conditionallyLogProgress(progressed uint) {
	if fm.verbose {
		if (fm.inputCount-progressed)%2000 > fm.inputCount%2000 {
			fmt.Fprintf(os.Stderr, "%d times runned: inspecting %d coverages\n", fm.inputCount, fm.inspectingHashNumbers)
		}
	} else {
		if (fm.inputCount-progressed)%(fm.trials/5) > fm.inputCount%(fm.trials/5) {
			fmt.Fprintf(os.Stderr, "%d times runned: inspecting %d coverages\n", fm.inputCount, fm.inspectingHashNumbers)
		}
	}
}

type endInfo struct {
	completeTrials bool
	maxedCAPACITY  bool
}

// 입력 개수를 검사하여 1) 요구받은 반복횟수를 다 채웠는지. 2) 퍼즈매니져의 배열이 들어찼는지 를 평가합니다.
func (fm *fuzzManager) evaluateEndState(updatedPriority uint, hn hashNumber) endInfo {
	if fm.inputCount >= fm.trials {
		return endInfo{
			completeTrials: true,
			maxedCAPACITY:  false,
		}
	} else if uint64(updatedPriority) >= (MaxCapacity-1) || uint64(hn) >= (MaxCapacity-1) {
		// Case 1: Priority reaches MAX. Prior increases discontinuously
		// Case 2 HashNumber reaches Max.
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

// 시드의 묶음들이 공유하는 '해시넘버'를 기반으로 중첩된 코디네이팅을 한번에 처리합니다.
// abstractNumber는 같은 해시넘버, 즉 커버리지를 공유하는 시드들이 몇 개인지를 나타냅니다.
// 시드를 저장하지 않고 추상적인 값만 배치함으로써, 연기 현상 및 고착화 현상을 제거할 수 있습니다.
// 이 함수는 그 공유하는 개수만큼 해시넘버의 우선순위를 낮춥니다. 그럼으로써 많이 등장한 시드는 다음번에 덜 선택됩니다.
func (fm *fuzzManager) coordinateAbstraction(hn hashNumber, abstractNumber uint) endInfo {
	fm.inputCount = fm.inputCount + abstractNumber
	oldPriority := fm.hashNumber2Priority[hn]
	fm.hashNumber2Priority[hn] = oldPriority + abstractNumber
	updatedPriority := fm.hashNumber2Priority[hn]

	if oldPriority == 0 {
		fm.inspectingHashNumbers++
		if fm.priorityCache > updatedPriority {
			fm.priorityCache = updatedPriority
		}
		fm.hashNumber2Seeds[hn] = newSeedQueue()
	} else {
		fm.priority2HashNumSet[oldPriority].remove(hn)
	}
	if fm.priority2HashNumSet[updatedPriority] == nil {
		fm.priority2HashNumSet[updatedPriority] = newHashNumberSet()
	}
	fm.priority2HashNumSet[updatedPriority].add(hn)
	fm.conditionallyLogProgress(abstractNumber)
	return fm.evaluateEndState(updatedPriority, hn)
}

// 앞선 코디네이팅 함수들은 자동적으로 o(1)시간 내로 시드와 해시넘버들을 우선순위에 맞게 정렬해 둡니다.
// 또한, 가장 최근의 우선순위가 얼마였는지를 prioiryCache에 기록해두어 선택 역시 o(1)시간 내에 완료할 수 있습니다(평균을 구할 시 o(1)이 됨).
func (fm *fuzzManager) popSeeds() []seed {
	var highestProrityHashNumber uint
	// 가장 높은, 그러니까 가장 0에 가까운 우선순위를 지니는 해시넘버를 찾습니다.
	for {
		hashNumberSet := fm.priority2HashNumSet[fm.priorityCache]
		if hashNumberSet == nil {
			fm.priorityCache++
			continue
		}
		hn, isExist := fm.priority2HashNumSet[fm.priorityCache].getAny()
		if !isExist {
			fm.priorityCache++
			continue
		}
		highestProrityHashNumber = uint(hn)
		break
	}
	// 해당 해시넘버(=커버리지에 매핑됨)를 지니는 시드를 꺼냅니다.
	popedSeed1, isOnce := fm.hashNumber2Seeds[highestProrityHashNumber].dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	peekedSeed2, err := fm.hashNumber2Seeds[highestProrityHashNumber].peek()
	if err != nil {
		return []seed{popedSeed1}
	}
	// 각 시드의 '부모'가 누구였는지를 검사합니다.
	pid1 := popedSeed1.pid
	pid2 := peekedSeed2.pid
	if pid1 == pid2 {
		// 같은 계통의 자손이여야 같이 pop되어 교배됩니다.
		popedSeed2, _ := fm.hashNumber2Seeds[highestProrityHashNumber].dequeue()
		return []seed{popedSeed1, popedSeed2}
	} else {
		// 서로 다른 계통의 자손들끼리는 같이 묶이지 않도록 합니다.
		// 그렇게 해야 고착화가 막아집니다.
		return []seed{popedSeed1}
	}
}

// 해시넘버를 바탕으로 그 해시넘버를 지니는 시드를 가져옵니다.
// 퍼져의 초기 포멧팅 과정에서, 해시넘버를 바탕으로 "최초의 시드들"을 가져올 때 씁니다.
func (fm *fuzzManager) popInitSeedByHN(hn hashNumber) seed {
	popedSeed, isOnce := fm.hashNumber2Seeds[hn].dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	return popedSeed
}

type (
	priorytyHashNumberList []priorytyAndHashNumber
	priorytyAndHashNumber  struct {
		priority   uint
		hashNumber int
	}
)

// 정렬을 구현하기 위해 구현해둔 매서드들 입니다.
func (phl priorytyHashNumberList) Len() int           { return len(phl) }
func (phl priorytyHashNumberList) Less(i, j int) bool { return phl[i].priority < phl[j].priority }
func (phl priorytyHashNumberList) Swap(i, j int)      { phl[i], phl[j] = phl[j], phl[i] }

type seedAndPrioryty struct {
	seed     seed
	priority uint
}

// 퍼즈매니져의 배열이 꽉 찬 경우, 우선순위가 높은 시드들 위주로 정보를 압축하는 함수입니다.
// 1) 같은 해시넘버(=커버리지)를 지니는 시드들을 하나로 압축합니다.
// 2) 압축된 시드의 정보를, 압축된 정도에 따라 변형합니다. 많이 압축될 수록 우선순위가 낮아집니다.
// 이렇게 하여 퍼즈매니져의 정보를 시드의 배열에 압축시킵니다.
func (fm *fuzzManager) summarize() []seed {
	// 현재 퍼즈매니져가 지닌 모든 (priority, hashnumber) 쌍을 구합니다. (각 hashnumber는 그에따른 priority를 지닙니다.)
	collectedPriorytyAndHN := priorytyHashNumberList{}
	for i, e := range fm.hashNumber2Priority {
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

	// (seed,priority)의 배열을 만듭니다.
	// 이때, seed: priority=n:1관계이지만, 시드 중 하나를 골라 이를 1:1페어로 만듭니다.
	sampledSeedAndPrioryty := []seedAndPrioryty{}

	for _, pair := range collectedPriorytyAndHN {
		hn := pair.hashNumber
		// 가장 앞선 시드를 샘플링합니다. 사실상 랜덤과 같습니다.
		seed := fm.hashNumber2Seeds[hn].data[0]
		priority := pair.priority
		sampledSeedAndPrioryty = append(sampledSeedAndPrioryty, seedAndPrioryty{seed, priority})
	}

	total := len(sampledSeedAndPrioryty)

	summarizedSeeds := []seed{}
	// 시드의 배열에 퍼즈매니져의 정보를 압축합니다.
	// 현재까지 해당 시드와 유사한 시드가 얼마나 나왔는지를 반영하여 seed.gen을 조정합니다.
	// seed.gen을 낮게 조정할수록, evlove함수에서 더 많은 값을 생산하게 됩니다.
	// 또한 더 일찌 pop됩니다.
	for i, seedAndPriority := range sampledSeedAndPrioryty {

		rank := i + 1

		seed := seedAndPriority.seed
		seed.pid = uint(i)
		seed.id = uint(i)
		seed.gen = allocateGen(seedAndPriority.priority, rank, total)
		seed.isCoordinated = false

		summarizedSeeds = append(summarizedSeeds, seed)
	}
	return summarizedSeeds
}

// 우선순위, 우선순위의 순위, 전체 우선순위의 종류를 받아, gen(희귀도와 유사)을 리턴합니다.
func allocateGen(priority uint, rank int, total int) uint {
	var bigger uint

	// 우선순위의 순위가 1등에 가까울수록 gen이 낮아집니다.
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
	c1CutLine := uint(float64(MaxCapacity) * 0.01)
	c2CutLine := uint(float64(MaxCapacity) * 0.02)
	c3CutLine := uint(float64(MaxCapacity) * 0.03)
	c4CutLine := uint(float64(MaxCapacity) * 0.05)
	c5CutLine := uint(float64(MaxCapacity) * 0.08)
	c6CutLine := uint(float64(MaxCapacity) * 0.1)
	c7CutLine := uint(float64(MaxCapacity) * 0.15)
	c8CutLine := uint(float64(MaxCapacity) * 0.30)
	// 희귀도가 높을수록 gen 낮음
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

func newFuzzManager(trials uint, inputCount uint) *fuzzManager {
	return &fuzzManager{
		inputCount: inputCount,
		trials:     trials,

		priorityCache: 1,

		hashNumber2Priority: &[MaxCapacity]uint{},
		hashNumber2Seeds:    &[MaxCapacity]*seedQueue{},
		priority2HashNumSet: &[MaxCapacity]hashNumberSet{},

		inspectingHashNumbers: 0,
	}
}

// 시드 큐
type (
	seedQueue struct {
		data []seed
	}
	seed struct {
		pid           uint
		id            uint
		gen           uint
		isCoordinated bool
		hashNumber    hashNumber
		content       []interface{}
		result        interface{}
	}
)

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

// hashNumberSet 타입 정의
type hashNumberSet map[hashNumber]struct{}

// NewHashNumberSet 생성자 함수
func newHashNumberSet() hashNumberSet {
	return make(hashNumberSet)
}

// Add: Set에 요소 추가
func (s hashNumberSet) add(hn hashNumber) {
	s[hn] = struct{}{}
}

// Remove: Set에서 요소 제거
func (s hashNumberSet) remove(hn hashNumber) {
	delete(s, hn)
}

// Contains: Set에 요소가 있는지 확인
func (s hashNumberSet) contains(hn hashNumber) bool {
	_, exists := s[hn]
	return exists
}

// GetAny: Set에서 임의의 요소 하나를 반환
func (s hashNumberSet) getAny() (hashNumber, bool) {
	for k := range s {
		return k, true
	}
	return 0, false
}

// Size: Set의 크기 반환
func (s hashNumberSet) size() int {
	return len(s)
}

// Clear: Set의 모든 요소 제거
func (s hashNumberSet) clear() {
	for key := range s {
		delete(s, key)
	}
}

// ToSlice: Set의 요소를 슬라이스로 변환
func (s hashNumberSet) toSlice() []hashNumber {
	keys := make([]hashNumber, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	return keys
}
