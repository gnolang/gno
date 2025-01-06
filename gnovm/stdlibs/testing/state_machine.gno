package testing

import (
	"sort"
	"strconv"
	"strings"
)

type SupportedType string

const (
	Byte_Array SupportedType = "[]byte"
	String     SupportedType = "string"
	Bool       SupportedType = "bool"
	Byte       SupportedType = "byte"
	Rune       SupportedType = "rune"
	Float32    SupportedType = "float32"
	Float64    SupportedType = "float64"
	Int        SupportedType = "int"
	Int8       SupportedType = "int8"
	Int16      SupportedType = "int16"
	Int32      SupportedType = "int32"
	Int64      SupportedType = "int64"
	Uint       SupportedType = "uint"
	Uint8      SupportedType = "uint8"
	Uint16     SupportedType = "uint16"
	Uint32     SupportedType = "uint32"
	Uint64     SupportedType = "uint64"
)

type Seed struct {
	Pid           uint
	Id            uint
	Gen           uint
	IsCoordinated bool
	HashNumber    HashNumber
	Content       []interface{}
	Result        interface{}
}

const (
	// Effective when path kinds(coverage) is in range 300,000~1,000,000.
	// and effiective when iters is less than 100_000_000
	// More than that is possible, but less effective.
	MAX_CAPACITY     uint64 = 100_000
	Partial_Capacity        = MAX_CAPACITY / 5
)

type StateMachine struct {
	Trials      uint
	Input_count uint

	All_Coverage     Coverage // []coveredline
	Covered_Coverage Coverage // []coveredline

	// for f.Add in f.Fuzz
	SeedType               []SupportedType
	init_hashnumber        []HashNumber
	string_byte_candidates []int

	// for f.Fuzz
	Priority_Cache         uint
	HashNumber2Seeds       *[MAX_CAPACITY]*Seed_Queue
	HashNumber2Priority    *[MAX_CAPACITY]uint
	Priority2HashNumberULL *[MAX_CAPACITY]*Unique_Uint_LinkedList

	Crash_Logger          *Crash_Logger
	verbose               bool
	inspectingHashNumbers uint
}

func New_StateMachine(trials uint, inputCount uint) *StateMachine {
	setted_AllULL := func() *[MAX_CAPACITY]*Unique_Uint_LinkedList {
		arr := &[MAX_CAPACITY]*Unique_Uint_LinkedList{}
		for i := range arr {
			arr[i] = New_Unique_Uint_LinkedList()
		}
		return arr
	}()
	return &StateMachine{
		Input_count:      inputCount,
		Trials:           trials,
		All_Coverage:     Coverage{{co_name: "nil", co_line: 0}},
		Covered_Coverage: Coverage{{co_name: "nil", co_line: 0}},

		Priority_Cache: 1,

		HashNumber2Priority:    &[MAX_CAPACITY]uint{},
		HashNumber2Seeds:       &[MAX_CAPACITY]*Seed_Queue{},
		Priority2HashNumberULL: setted_AllULL,

		inspectingHashNumbers: 0,
		Crash_Logger:          New_Crash_Logger(),
	}
}

type EndInfo struct {
	Complete_Trials bool
	MAXed_CAPACITY  bool
}

func (sm *StateMachine) CoordinateAbstraction(hn HashNumber, abstract_number uint) EndInfo {
	sm.Input_count = sm.Input_count + abstract_number
	// 현재의 priort를 알아올 수 있음
	// 1. prior는 불연속 증가함
	// 2. hn는 반드시 연속 증가함
	sm.HashNumber2Priority[hn] = sm.HashNumber2Priority[hn] + abstract_number
	next_priority := sm.HashNumber2Priority[hn]
	next_hn := hn
	if sm.verbose {
		if (sm.Input_count-abstract_number)%2000 > sm.Input_count%2000 {
			println(sm.Input_count, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	} else {
		if (sm.Input_count-abstract_number)%(sm.Trials/5) > sm.Input_count%(sm.Trials/5) {
			println(sm.Input_count, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	}
	if sm.Input_count >= sm.Trials {
		return EndInfo{
			Complete_Trials: true,
			MAXed_CAPACITY:  false,
		}
	} else if uint64(next_priority) >= (MAX_CAPACITY-1) || uint64(next_hn) >= (MAX_CAPACITY-1) {
		// Case 1: Priority reaches MAX (and protect overflow)
		// Case 2 HashNumber reaches Max (and protect overflow)
		return EndInfo{
			Complete_Trials: false,
			MAXed_CAPACITY:  true,
		}
	} else {
		return EndInfo{
			Complete_Trials: false,
			MAXed_CAPACITY:  false,
		}
	}
}

func (sm *StateMachine) CoordinateSeed(seed Seed) EndInfo {
	if seed.IsCoordinated {
		sm.HashNumber2Seeds[seed.HashNumber].Enqueue(seed)
		return EndInfo{
			Complete_Trials: false,
			MAXed_CAPACITY:  false,
		}
	}
	hn := seed.HashNumber
	seed.IsCoordinated = true
	sm.Input_count++

	if sm.HashNumber2Seeds[hn] == nil {
		sm.inspectingHashNumbers++
		sm.HashNumber2Seeds[hn] = New_Seed_Queue()
	}
	sm.HashNumber2Seeds[hn].Enqueue(seed)
	old_priority := sm.HashNumber2Priority[hn]
	sm.HashNumber2Priority[hn]++
	updated_priority := sm.HashNumber2Priority[hn]

	if updated_priority == 1 {
		sm.Priority_Cache = 1
		sm.Priority2HashNumberULL[updated_priority].Append(uint(hn))
	} else {
		sm.Priority2HashNumberULL[old_priority].Delete(uint(hn))
		sm.Priority2HashNumberULL[updated_priority].Append(uint(hn))
	}
	if sm.verbose {
		if sm.Input_count%2000 == 0 {
			println(sm.Input_count, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	} else {
		if sm.Input_count%(sm.Trials/5) == 0 {
			println(sm.Input_count, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	}

	if sm.Input_count >= sm.Trials {
		return EndInfo{
			Complete_Trials: true,
			MAXed_CAPACITY:  false,
		}
	} else if uint64(updated_priority) >= (MAX_CAPACITY-1) || uint64(hn) >= (MAX_CAPACITY-1) {
		// Case 1: Priority reaches MAX
		// Case 2 HashNumber reaches Max
		return EndInfo{
			Complete_Trials: false,
			MAXed_CAPACITY:  true,
		}
	} else {
		return EndInfo{
			Complete_Trials: false,
			MAXed_CAPACITY:  false,
		}
	}
}

func (sm *StateMachine) PopInitSeedByHN(hn HashNumber) Seed {
	popedSeed, isOnce := sm.HashNumber2Seeds[hn].Dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	return popedSeed
}

func (sm *StateMachine) PopSeeds() []Seed {
	var highest_hn uint

	for {

		hn, isExist := sm.Priority2HashNumberULL[sm.Priority_Cache].Peek()
		if !isExist {
			sm.Priority_Cache++

			continue
		}

		highest_hn = hn
		break
	}

	popedSeed_1, isOnce := sm.HashNumber2Seeds[highest_hn].Dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	peekedSeed_2, err := sm.HashNumber2Seeds[highest_hn].Peek()
	if err != nil {
		return []Seed{popedSeed_1}
	}
	pid_1 := popedSeed_1.Pid
	pid_2 := peekedSeed_2.Pid
	if pid_1 == pid_2 {
		popedSeed_2, _ := sm.HashNumber2Seeds[highest_hn].Dequeue()
		return []Seed{popedSeed_1, popedSeed_2}
	} else {
		return []Seed{popedSeed_1}
	}
}

type P_HN_Pair struct {
	Priority   uint
	HashNumber int
}
type PHP_List []P_HN_Pair

func (phpl PHP_List) Len() int           { return len(phpl) }
func (phpl PHP_List) Less(i, j int) bool { return phpl[i].Priority < phpl[j].Priority }
func (phpl PHP_List) Swap(i, j int)      { phpl[i], phpl[j] = phpl[j], phpl[i] }

type S_P_Pair struct {
	Seed     Seed
	Priority uint
}

func (sm *StateMachine) Summarize() []Seed {
	collected_P_HN_Pair := PHP_List{}
	for i, e := range sm.HashNumber2Priority {

		if e == 0 {
			break
		}
		collected_P_HN_Pair = append(collected_P_HN_Pair, P_HN_Pair{
			Priority:   e,
			HashNumber: i,
		})
		// 슬라이스 길이가 Partial_Capacity를를 초과하면 종료
		if len(collected_P_HN_Pair) >= int(Partial_Capacity) {
			break
		}
	}

	sort.Sort(collected_P_HN_Pair)

	sampled_sp_pair := []S_P_Pair{}

	for _, pair := range collected_P_HN_Pair {
		hn := pair.HashNumber
		seed := sm.HashNumber2Seeds[hn].data[0]
		priority := pair.Priority
		sampled_sp_pair = append(sampled_sp_pair, S_P_Pair{seed, priority})
	}

	total := len(sampled_sp_pair)

	summarized_seeds := []Seed{}

	for i, sp_pair := range sampled_sp_pair {

		rank := i + 1

		seed := sp_pair.Seed
		seed.Pid = 0
		seed.Id = uint(i)
		seed.Gen = genAllocation(sp_pair.Priority, rank, total)
		seed.IsCoordinated = false

		summarized_seeds = append(summarized_seeds, seed)
	}
	return summarized_seeds
}

func genAllocation(priority uint, rank int, total int) uint {
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
	c1_cutLine := uint(float64(MAX_CAPACITY) * 0.008)
	c2_cutLine := uint(float64(MAX_CAPACITY) * 0.014)
	c3_cutLine := uint(float64(MAX_CAPACITY) * 0.02)
	c4_cutLine := uint(float64(MAX_CAPACITY) * 0.03)
	c5_cutLine := uint(float64(MAX_CAPACITY) * 0.08)
	c6_cutLine := uint(float64(MAX_CAPACITY) * 0.1)
	c7_cutLine := uint(float64(MAX_CAPACITY) * 0.15)
	c8_cutLine := uint(float64(MAX_CAPACITY) * 0.30)
	switch {
	case priority <= c1_cutLine:
		if bigger > 1 {
			bigger = 1
		}
	case priority <= c2_cutLine:
		if bigger > 2 {
			bigger = 2
		}
	case priority <= c3_cutLine:
		if bigger > 3 {
			bigger = 3
		}
	case priority <= c4_cutLine:
		if bigger > 4 {
			bigger = 4
		}
	case priority <= c5_cutLine:
		if bigger > 5 {
			bigger = 5
		}
	case priority <= c6_cutLine:
		if bigger > 6 {
			bigger = 6
		}
	case priority <= c7_cutLine:
		if bigger > 7 {
			bigger = 7
		}
	case priority <= c8_cutLine:
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
