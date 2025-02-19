package fuzzing_test

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/fuzzing"
	"unicode/utf8"
)

// TestReflectMutationInfo_RootEnt verifies that RootEnt's ReflectMutationInfo
// correctly updates mutation strategy counts and entity scores.
func TestReflectMutationInfo_RootEnt(t *testing.T) {
	// Create RootEnt
	root := fuzzing.ParseToRoot([]byte("hello world"))
	root.InitStrategyCount()

	// Create MutationInfo for RootEnt
	// Example: Assume DeleteStart (DS) is used twice, and ChangeRandomly (CHR) is used once.
	mutInfo := fuzzing.MutationInfo{
		IsMutated:           true,
		IsStringOrByteArray: true,
		EntId:               2,
		UsedStrategyMap: map[fuzzing.MutationCode]int{
			fuzzing.DS:  2,
			fuzzing.CHR: 1,
		},
		UsedCodesForNotSbMap: nil,
	}

	// Call ReflectMutationInfo: Updates the strategy count and Score of child entities.
	root.ReflectMutationInfo(mutInfo)

	// Expected outcome: root.Score should increase by 1 + 0 = 0.
	if root.Score != 1 {
		t.Errorf("Expected root Score to be %d, got %d", 1, root.Score)
	}

	// Expected value for each strategy: Initial value 1 + (usage count × 5).
	expectedScore := 1 + 1*5
	expectedDS := 1 + 2*5  // 1+10 = 11
	expectedCHR := 1 + 1*5 // 1+5 = 6
	// The ID of the modified entity.
	ent := root.Id2Ent[2]
	switch e := ent.(type) {
	case *fuzzing.LetterEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	case *fuzzing.NumEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	case *fuzzing.WsEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	case *fuzzing.MarkEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	case *fuzzing.KeyValEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	case *fuzzing.GroupEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	case *fuzzing.RootEnt:
		verifyStrategyCounts(t, e.Score, e.Strategy2count, expectedDS, expectedCHR, expectedScore)
	default:
		t.Errorf("Unexpected type for root.Id2Ent[2]: %T", ent)
	}
}

// verifyStrategyCounts is a helper function to validate strategy count updates
// for different entity types.
func verifyStrategyCounts(t *testing.T, score uint64, strategy2count map[fuzzing.MutationCode]uint64, expectedDS, expectedCHR, expectedScore int) {
	if score != uint64(expectedScore) {
		t.Errorf("Expected Score to be %d, got %d", expectedScore, score)
	}
	if strategy2count[fuzzing.DS] != uint64(expectedDS) {
		t.Errorf("Expected DS count to be %d, got %d", expectedDS, strategy2count[fuzzing.DS])
	}
	if strategy2count[fuzzing.CHR] != uint64(expectedCHR) {
		t.Errorf("Expected CHR count to be %d, got %d", expectedCHR, strategy2count[fuzzing.CHR])
	}
}

// TestReflectMutationInfo_EntForNotSB verifies that EntForNotSB correctly updates
// mutation counts for non-string and non-byte entities.
func TestReflectMutationInfo_EntForNotSB(t *testing.T) {
	// Example EntForNotSB creation (integer 42).
	ens := fuzzing.ParseToEntForNotSB(42)
	ens.InitStrategyCount() // All strategy counts are initialized to 1.

	// Create MutationInfo for NotSb.
	// Example: Assume SmallPlus (SmlPls) is used three times, and SmallMinus (SmlMns) is used twice.
	mutInfo := fuzzing.MutationInfo{
		IsMutated:           true,
		IsStringOrByteArray: false, // Since it is an integer type, not string or []byte type

		UsedStrategyMap: nil,
		UsedCodesForNotSbMap: map[fuzzing.MutationCodeForNotSB]int{
			fuzzing.SmlPls: 3,
			fuzzing.SmlMns: 2,
		},
	}

	// Call ReflectMutationInfo (method specialized for EntForNotSB).
	ens.ReflectMutationInfo(mutInfo)

	// Expected: Each strategy count is initialized to 1 + (number of uses × 5).
	expectedSmlPls := uint64(1 + 3*5) // 16
	expectedSmlMns := uint64(1 + 2*5) // 11

	if ens.Strategy2count[fuzzing.SmlPls] != expectedSmlPls {
		t.Errorf("Expected SmlPls count to be %d, got %d", expectedSmlPls, ens.Strategy2count[fuzzing.SmlPls])
	}
	if ens.Strategy2count[fuzzing.SmlMns] != expectedSmlMns {
		t.Errorf("Expected SmlMns count to be %d, got %d", expectedSmlMns, ens.Strategy2count[fuzzing.SmlMns])
	}
}

// TestMutateArg repeatedly calls MutateArg on different argument types
// to verify mutation behavior and randomness.
func TestMutateArg(t *testing.T) {
	seed := fuzzing.Seed{
		Content:    []interface{}{100, true, 3.14, "H"},
		Id:         1,
		Pid:        0,
		Gen:        0,
		HashNumber: 123,
	}

	ms := fuzzing.NewMutSample(seed)

	for idx, original := range seed.Content {
		t.Logf("original args [idx=%d]: %v", idx, original)
		for iter := 0; iter < 5; iter++ {
			mutatedArg, mutationInfo := ms.MutateArg(original, 3, idx)
			t.Logf("  Iteration %d: mutated value: %v, mutation info: %+v", iter, mutatedArg, mutationInfo)
		}
	}
}

// TestSort tests boundary-based selection logic for mappedGeneratingNumber.
func TestSort(t *testing.T) {
	boundarySamplePopularity := []uint64{2, 8, 32, 128, 512, 2048, 8096, 16192}
	mappedGeneratingNumber := []int{1440, 1080, 720, 600, 480, 360, 240, 120}

	testCases := []uint64{1, 8, 30, 20_000, 1231232121}

	for _, e := range testCases {
		idx := sort.Search(len(boundarySamplePopularity), func(i int) bool {
			return e <= boundarySamplePopularity[i]
		})

		var generatingNumber int
		if idx < len(boundarySamplePopularity) {
			generatingNumber = mappedGeneratingNumber[idx]
		} else {
			generatingNumber = 120
		}

		t.Logf("e: %d, selected: %d", e, generatingNumber)
	}
}

func TestParser(t *testing.T) {
	fmt.Println("=== Final AST ===")

	program := `func main() {
 		parser := NewParser(input)
 		t1 := parser.parseTokens(0)
 		root := rootEnt{"depth": '-0x12', data: 0X}
 		root = applyPatternToRoot(root)
 		printTokens([]Ent{root}, 0)
 		printTokens([]Ent{root}, 0)
 		}`

	root := fuzzing.ParseToRoot([]byte(fuzzing.SafeUTF8(program)))
	fuzzing.PrintTree(root, 0)
	myAst := fuzzing.ParseToRoot([]byte(program))

	println("Root has", len(myAst.Id2Ent), "nodes total.")

	// Directly accesses an entity by its specific ID.
	if ent, ok := myAst.Id2Ent[16]; ok {
		fmt.Println("ID=16 is Ent:", ent.String())
	}
	pp := "111/11\xeb"
	rootp := fuzzing.ParseToRoot([]byte(pp))
	fuzzing.PrintTree(rootp, 0)
	start, end := fuzzing.SliceWithOutID(1, rootp)
	println("start:", fuzzing.EscapedToString(start), "end:", fuzzing.EscapedToString(end))
	ml := `<html><a>yeah!<a/><html/>`
	rootm := fuzzing.ParseToRoot([]byte(ml))
	fuzzing.PrintTree(rootm, 0)
}
