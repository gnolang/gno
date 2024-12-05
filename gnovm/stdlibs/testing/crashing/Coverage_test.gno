package crashing

import "testing"

func Test_Is_exist_coverage(t *testing.T) {
	// 미리 정의된 set 만들어놓기
	coverageSet := NewSet()

	testVal1 := []uint{1, 11, 1337}
	testVal2 := []uint{1, 143, 1337}
	testVal3 := []uint{9832, 10, 1337}

	testVal1ST := UintToString(testVal1)
	testVal2ST := UintToString(testVal2)
	testVal3ST := UintToString(testVal3)

	exVal1 := UseSum64String(testVal1ST)
	exVal2 := UseSum64String(testVal2ST)
	exVal3 := UseSum64String(testVal3ST)

	coverageSet.Add(exVal1)
	coverageSet.Add(exVal2)
	coverageSet.Add(exVal3)

	// 테스트용 Val 생성
	testVal := []uint{1, 10, 1337}
	testValST := UintToString(testVal)
	testValHash := UseSum64String(testValST)

	println(testValHash)

	// test
	println(coverageSet.Contains(testValHash))
}
