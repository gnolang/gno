package fuzzinglib

import (
	"math/rand"
	"testing"
)

// TODO: 수정하기
func randomByte() rune {
	r := int(testing.RandRange(32, 96))
	return rune(r)
}

const (
	ABSOLUTE_MIN  = 0x00
	SPECIAL_MAX   = 0x1F
	LOW_ASCII_MIN = 0x20
	LOW_ASCII_MAX = 0x7E

	HIGH_ASCII_MIN = 0x80
	HIGH_ASCII_MAX = 0xFF

	PRINT_UNICODE_MIN = 0x100
	PRINT_UNICODE_MAX = 0xD7FF

	BOUNDARY_UNICODE_MIN = 0xD800
	BOUNDARY_UNICODE_MAX = 0x10FFFF

	ABSOLUTE_MAX = 0x7FFFFFFF
)

// 수정된 randomRune 함수
func randomRune_from(seed_rune rune) rune {
	p := rand.Float64() // 0.0 <= p < 1.0

	var r int64
	// 현재 케이스를 판별
	currentCase := determineCase(seed_rune)

	// 확률에 따른 케이스 이동
	// 점진적인

	switch {
	case p < 0.3: // 이전 케이스
		currentCase = (currentCase + 5) % 6 // 순환 처리
		break
	case p < 0.7: // 본인 케이스
		break
		// 유지
	case p <= 1.0: // 다음 케이스
		currentCase = (currentCase + 1) % 6 // 순환 처리
	}

	// 해당 케이스의 범위에서 랜덤 값 생성
	switch currentCase {
	case 0: // 0~31 스페셜 문자
		r = testing.RandRange(ABSOLUTE_MIN, SPECIAL_MAX)
	case 1: // 32~126 아스키
		r = testing.RandRange(LOW_ASCII_MIN, LOW_ASCII_MAX)
	case 2: // 128~255 하이 아스키
		r = testing.RandRange(HIGH_ASCII_MIN, HIGH_ASCII_MAX)
	case 3: // 그 외 프린트 가능한 유니코드
		r = testing.RandRange(PRINT_UNICODE_MIN, PRINT_UNICODE_MAX)
	case 4: // 그 외의 프린트 불가능한 유니코드
		r = testing.RandRange(BOUNDARY_UNICODE_MIN, BOUNDARY_UNICODE_MAX)
	case 5:
		// 유니코드가 아닌 케이스
		r = testing.RandRange(BOUNDARY_UNICODE_MAX, ABSOLUTE_MAX)
	}

	return rune(r)
}

// determineCase: seed_rune의 케이스를 판별
// 유니코드를 점진적으로 변형시키기.
func determineCase(seed_rune rune) int {
	switch {
	// 00~31의 특수문자 케이스
	case seed_rune >= ABSOLUTE_MIN && seed_rune <= SPECIAL_MAX:
		return 0
	// 32 ~127의 아스키 케이스
	case seed_rune >= LOW_ASCII_MIN && seed_rune <= LOW_ASCII_MAX:
		return 1
	// 128 ~256의 높은 아스키 케이스
	case seed_rune >= HIGH_ASCII_MIN && seed_rune <= HIGH_ASCII_MAX:
		return 2
	// 프린트 가능한 유니코드 케이스
	case seed_rune >= PRINT_UNICODE_MIN && seed_rune <= PRINT_UNICODE_MAX:
		return 3
	// 프린트 불가능한 유니코드 케이스
	case seed_rune >= BOUNDARY_UNICODE_MIN && seed_rune <= BOUNDARY_UNICODE_MAX:
		return 4
	// 유니코드가 아닌 케이스
	default:
		return 5
	}
}

// func main() {
// 	seed1 := Seed{Gen: 1, IsUsed: false, Content: []interface{}{"apple"}, Result: nil}
// 	// seed2 := Seed{Gen: 3, IsUsed: false, Content: []interface{}{"wellcome!"}, Result: nil}

// 	old_p := seed1
// 	for i := 0; i < 10; i++ {
// 		new_gen := evolve([]Seed{old_p})
// 		println(i+1, "번째 변이 결과-1:", new_gen[0].Content)
// 		println(i+1, "번째 변이 결과-2:", new_gen[1].Content)
// 		old_p = new_gen[0]
// 	}

// 	// a := 'a'
// 	// for i := 0; i < 10; i++ {
// 	// 	a = rune(a)
// 	// 	a = randomRune_from(a)
// 	// }
// }

// Seed의 Content를 랜덤으로 변형
// 다변함수 시엔 하나를 잡아서 변형형
// TODO: 다른 타입 가능하게 하기기
// TODO: 변형 전략 강화도 고려하기. 일단 테스트 후, 좀더 강한 변형도 고려하기.
// TODO: 그래도, 이런 변형이 나을지도. 점진적이니까.
func Mutate(seed Seed) Seed {
	if len(seed.Content) == 0 {
		return seed
	}
	index := 0
	if len(seed.Content) > 1 {
		index = int(testing.RandRange(0, len(seed.Content)-1))
	}

	selected := seed.Content[index]
	// 선택된 원소를 수정
	str, ok := selected.(string)
	if !ok {
		return seed
	}

	runes := []rune(str)
	if len(runes) > 0 {
		runeIndex := testing.RandRange(0, len(runes)-1)
		runes[runeIndex] = randomRune_from(runes[runeIndex])
	}
	var new_str string = string(runes)
	seed.Content[index] = new_str

	return seed
}

// TODO: 다른 타입도 변형 가능하게 하기.
func InsertDelete(seed Seed, p float64) Seed {
	if len(seed.Content) == 0 {
		return seed
	}

	index := 0
	if len(seed.Content) > 1 {
		index = int(testing.RandRange(0, len(seed.Content)-1))
	}

	selected := seed.Content[index]
	// 선택된 원소를 수정
	str, ok := selected.(string)
	if !ok {
		return seed
	}

	rr := []rune(str)
	l := len(rr)

	// Insert
	if testing.GenerateRandomBool(p) {
		// 삽입할 문자 결정
		sample := rr[testing.RandRange(0, l-1)]
		char := randomRune_from(sample)
		// 삽입 포지션 배정
		pos := testing.RandRange(0, l-1)
		// 한 칸 늘리기
		rr = append(rr, 0)
		// 한 칸 밀기
		copy(rr[pos+1:], rr[pos:])
		// 그 사이에 삽입
		rr[pos] = char
	} else {
		if l == 0 {
			return seed
		}

		pos := testing.RandRange(0, l-1)
		rr = append(rr[:pos], rr[pos+1:]...)
	}
	var new_str string = string(rr)
	seed.Content[index] = new_str

	return seed
}

// 기존의 적합도, 피트니스를 삭제했습니다.
// AFl에 통합해서 속도를 높이려다 보니 그냥 큐, 스택, 링크드 리스트로 관리하는게 빠르다 판단했습니다.
// (제가 본 afl로직을 따라가면서 해당 피트니스 관리를 유지하게되면 뭔가 불편해집니다.)
// 피트니스, 선택 로직은 치환된 것이라 보면 되겠습니다.

// 기존의 교배 로직을 수정했습니다.
// 고착화 문제 해결을 위해 gen에 따른 수를 조정했습니다.
// 다중 교차점 로직으로 바꿨습니다.

// TODO: 문자열 이외도 가능케 하기
// TODO: 숫자 교배는 훨씬 보수적으로.
func TwoPointCrossover(parent1, parent2 Seed) (Seed, Seed) {
	// 깊은 복사를 위해 새로운 슬라이스 생성
	content1 := make([]interface{}, len(parent1.Content))
	for i, v := range parent1.Content {
		content1[i] = v // 안전하게 string으로 캐스팅
	}
	content2 := make([]interface{}, len(parent2.Content))
	for i, v := range parent2.Content {
		content2[i] = v
	}

	for i := 0; i < len(parent1.Content); i++ {
		str1, ok1 := parent1.Content[i].(string)
		if !ok1 {
			continue
		}
		str2, ok2 := parent2.Content[i].(string)
		if !ok2 {
			continue
		}
		p1Runes := []rune(str1)
		p2Runes := []rune(str2)

		p1Len := len(p1Runes)
		p2Len := len(p2Runes)

		// 최소 길이를 기준으로 교배 지점 설정
		minLen := p1Len
		if p2Len < p1Len {
			minLen = p2Len
		}

		point1 := testing.RandRange(0, minLen)
		point2 := testing.RandRange(0, minLen)

		// 교차점 정렬 (point1 < point2 보장)
		if point1 > point2 {
			point1, point2 = point2, point1
		}
		// 자식 생성
		crossed_str1 := append([]rune{}, p1Runes[:point1]...)          // 부모1의 첫 구간
		crossed_str1 = append(crossed_str1, p2Runes[point1:point2]...) // 부모2의 중간 구간
		crossed_str1 = append(crossed_str1, p1Runes[point2:]...)       // 부모1의 마지막 구간

		crossed_str2 := append([]rune{}, p2Runes[:point1]...)          // 부모2의 첫 구간
		crossed_str2 = append(crossed_str2, p1Runes[point1:point2]...) // 부모1의 중간 구간
		crossed_str2 = append(crossed_str2, p2Runes[point2:]...)       // 부모2의 마지막 구간

		result_str1 := string(crossed_str1)
		result_str2 := string(crossed_str2)
		content1[i] = result_str1
		content2[i] = result_str2
	}

	// 새로운 자식 생성
	updatedIdv1 := Seed{Gen: parent1.Gen + 1, IsUsed: false, Content: content1, Result: nil}
	updatedIdv2 := Seed{Gen: parent2.Gen + 1, IsUsed: false, Content: content2, Result: nil}

	return updatedIdv1, updatedIdv2
}

// parents를 받아 children을 리턴합니다.
// 교배의 수와 정도를 gen따라 유동화 했습니다
// 전체적으로 수정했습니다.
func evolve(seeds []Seed) []Seed {
	p1 := seeds[0]
	// 이게 목표 수
	target_count := 120 / (p1.Gen * p1.Gen)
	loop_count := target_count / 2

	// 결과가 2 미만일 경우 2로 설정
	if target_count < 2 {
		target_count = 2
	}
	if loop_count < 2 {
		loop_count = 2
	}

	// 결과를 짝수로 만듦
	if loop_count%2 != 0 {
		loop_count++
	}

	new_generation := []Seed{}

	if len(seeds) == 1 {
		// seed가 하나인 경우에도 evlove가능하게 디자인
		for i := 0; i < int(target_count); i++ {
			// 교배가 무의미하므로, "반드시" 둘 중 하나의 변형이 일어나야 함.

			// 깊은 복사를 위해 새로운 슬라이스 생성
			new_content := make([]interface{}, len(seeds[0].Content))
			for i, v := range seeds[0].Content {
				new_content[i] = v // 안전하게 string으로 캐스팅
			}

			new_ind := Seed{
				Gen:     seeds[0].Gen + 1,
				IsUsed:  false,
				Content: new_content,
				Result:  nil,
			}

			if testing.UniformRandomBool(0.6) {
				new_ind = Mutate(new_ind)
			} else {
				new_ind = InsertDelete(new_ind, 0.3)
			}
			new_ind.Gen = seeds[0].Gen + 1
			new_generation = append(new_generation, new_ind)

		}
		return new_generation
	}

	if len(seeds) > 3 {
		panic("not covered len")
	}

	// 두개인 경우 정상 유전 알고리즘
	p2 := seeds[1]
	// children 생성합니다. (120/(gen**2))
	for i := 0; i < int(loop_count); i++ {
		c1, c2 := TwoPointCrossover(p1, p2)
		new_generation = append(new_generation, c1)
		new_generation = append(new_generation, c2)

	}

	// 변이 확률은 기존 퍼징에 기반합니다.
	for i := range new_generation {

		if testing.UniformRandomBool(0.3) {
			new_generation[i] = Mutate(new_generation[i]) // 원본 데이터 수정
		}
		if testing.UniformRandomBool(0.2) {
			new_generation[i] = InsertDelete(new_generation[i], 0.3)
		}

	}

	return new_generation
}
