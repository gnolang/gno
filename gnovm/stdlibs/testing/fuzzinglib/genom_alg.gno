package fuzzinglib

import (
	"errors"
	"math"
	"testing"
	"time"
)

// func main() {
// 	var s1 string = "apple"
// 	var b1 []byte = []byte("aaaaaa")

// 	var s2 string = "banana"
// 	var b2 []byte = []byte("bbbbbb")

// 	seed1 := Seed{Gen: 3, IsUsed: false, Content: []interface{}{s1, b1}, Result: nil}
// 	seed2 := Seed{Gen: 3, IsUsed: false, Content: []interface{}{s2, b2}, Result: nil}

// 	// old_p := seed1

// 	// for i := 0; i < 10; i++ {
// 	// 	new_gen := evolve([]Seed{old_p})
// 	// 	println(i+1, "번째 변이 결과-1:", new_gen[0].Content)
// 	// 	println(i+1, "번째 변이 결과-2:", new_gen[1].Content)
// 	// 	old_p = new_gen[0]
// 	// }

// 	new_gen := evolve([]Seed{seed1, seed2})
// 	for i, e := range new_gen {
// 		println(i, "번쨰:", e.Content)
// 		v, ok := e.Content[1].([]byte)
// 		if !ok {
// 			panic("AAA")
// 		}
// 		println(string(v))
// 	}
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
		index = int(testing.RandRange(0, int64(len(seed.Content)-1)))
	}

	selected := seed.Content[index]
	// 선택된 원소를 수정
	switch v := selected.(type) {
	case int, int8, int16, int32, int64:
		seed.Content[index] = randomInt_from(v)
	case uint, uint8, uint16, uint32, uint64:
		seed.Content[index] = randomUint_from(v)
	case float32, float64:
		seed.Content[index] = randomFloat_from(v)
	case bool:
		seed.Content[index] = randomBool()
	case string:
		runes := []rune(v)
		if len(runes) > 0 {
			runeIndex := testing.RandRange(0, int64(len(runes)-1))
			runes[runeIndex] = randomRune_from(runes[runeIndex])
		}
		var new_str string = string(runes)
		seed.Content[index] = new_str
	case []byte:
		bytes := []byte(v)
		if len(bytes) > 0 {
			byteIndex := testing.RandRange(0, int64(len(bytes)-1))
			bytes[byteIndex] = randomByte_from(bytes[byteIndex])
		}
		var new_byt []byte = []byte(bytes)
		seed.Content[index] = new_byt
	default:
		panic("not supported type")
	}

	return seed
}

// TODO: 다른 타입도 변형 가능하게 하기.
func InsertDelete(seed Seed, p float64) Seed {
	if len(seed.Content) == 0 {
		return seed
	}

	index := 0
	if len(seed.Content) > 1 {
		index = int(testing.RandRange(0, int64(len(seed.Content)-1)))
	}

	selected := seed.Content[index]
	switch v := selected.(type) {
	case string:
		rr := []rune(v)
		l := len(rr)

		// Insert
		if testing.GenerateRandomBool(p) {
			// 삽입할 문자 결정
			sample := rr[testing.RandRange(0, int64(l-1))]
			char := randomRune_from(sample)
			// 삽입 포지션 배정
			pos := testing.RandRange(0, int64(l-1))
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

			pos := testing.RandRange(0, int64(l-1))
			rr = append(rr[:pos], rr[pos+1:]...)
		}
		var new_str string = string(rr)
		seed.Content[index] = new_str
	case []byte:
		bb := []byte(v)
		l := len(bb)

		// Insert
		if testing.GenerateRandomBool(p) {
			// 삽입할 문자 결정
			sample := bb[testing.RandRange(0, int64(l-1))]
			bt := randomByte_from(sample)
			// 삽입 포지션 배정
			pos := testing.RandRange(0, int64(l-1))
			// 한 칸 늘리기
			bb = append(bb, 0)
			// 한 칸 밀기
			copy(bb[pos+1:], bb[pos:])
			// 그 사이에 삽입
			bb[pos] = bt
		} else {
			if l == 0 {
				return seed
			}

			pos := testing.RandRange(0, int64(l-1))
			bb = append(bb[:pos], bb[pos+1:]...)
		}
		var new_byt []byte = []byte(bb)
		seed.Content[index] = new_byt
	default:

	}
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
		switch v1 := content1[i].(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			content1[i], content2[i] = linear_crossover(v1, content2[i])
		case bool:
			content1[i] = v1
			content2[i] = content2[i]
		case string:
			str1 := v1
			str2, ok2 := parent2.Content[i].(string)
			if !ok2 {
				panic("type not equal")
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

			point1 := testing.RandRange(0, int64(minLen))
			point2 := testing.RandRange(0, int64(minLen))

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
		case []byte:
			byt1 := v1
			byt2, ok := parent2.Content[i].([]byte)
			if !ok {
				panic("type not equal")
			}
			p1Bytes := []byte(byt1)
			p2Bytes := []byte(byt2)
			p1Len := len(p1Bytes)
			p2Len := len(p2Bytes)
			minLen := p1Len
			if p2Len < p1Len {
				minLen = p2Len
			}

			point1 := testing.RandRange(0, int64(minLen))
			point2 := testing.RandRange(0, int64(minLen))
			// 교차점 정렬 (point1 < point2 보장)
			if point1 > point2 {
				point1, point2 = point2, point1
			}
			// 자식 생성
			crossed_byt1 := append([]byte{}, p1Bytes[:point1]...)          // 부모1의 첫 구간
			crossed_byt1 = append(crossed_byt1, p2Bytes[point1:point2]...) // 부모2의 중간 구간
			crossed_byt1 = append(crossed_byt1, p1Bytes[point2:]...)       // 부모1의 마지막 구간

			crossed_byt2 := append([]byte{}, p2Bytes[:point1]...)          // 부모2의 첫 구간
			crossed_byt2 = append(crossed_byt2, p1Bytes[point1:point2]...) // 부모1의 중간 구간
			crossed_byt2 = append(crossed_byt2, p2Bytes[point2:]...)       // 부모2의 마지막 구간

			result_byt1 := []byte(crossed_byt1)
			result_byt2 := []byte(crossed_byt2)
			content1[i] = result_byt1
			content2[i] = result_byt2
		default:
			panic("not supported type")

		}
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
	target_count := 240 / (p1.Gen * p1.Gen)
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
				new_content[i] = v
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

// 랜덤 바이트를 점진적 변형
func randomByte_from(seed_byte byte) byte {
	p := testing.GetSingleRand().Float64() // 0.0 <= p < 1.0
	// 단지, del확률 처리용으로 만든 변수. 신경 쓸 필요 없음.
	next_flag := testing.GenerateRandomBool(0.5)
	var b uint8
	// 현재 케이스를 판별
	// 이때 케이스는 0~3의 4개
	currentCase := determineCase(rune(seed_byte))

	// 확률에 따른 케이스 이동
	// 점진적인
	if currentCase == 2 {
		// DEL케이스인 경우 확률 조정
		switch {
		case p < 0.45: // 이전 케이스
			currentCase = (currentCase + 3) % 4 // 순환 처리
			break
		case p < 0.55: // 본인 케이스 유지 확률은 10%
			break
		// 유지
		case p <= 1.0: // 다음 케이스
			currentCase = (currentCase + 1) % 4 // 순환 처리
		}
	} else {
		switch {
		case p < 0.15:
			currentCase = (currentCase + 2) % 4 // 순환 처리
			next_flag = false
			break
		case p < 0.35: // 이전 케이스
			currentCase = (currentCase + 3) % 4 // 순환 처리
			next_flag = false
			break
		case p < 0.65: // 본인 케이스
			break
			// 유지
		case p < 0.85:
			next_flag = true
			currentCase = (currentCase + 1) % 4 // 순환 처리
		case p <= 1.0: // 다음 케이스
			next_flag = true
			currentCase = (currentCase + 2) % 4 // 순환 처리
		}
	}

	if currentCase == 2 {
		p2 := testing.GetSingleRand().Float64()
		// DEL케이스인 경우 확률 조정
		if next_flag {
			if p2 < 0.8 {
				// 다음 케이스로 넘어왔는데 DEL인 경우
				// 높은 확률로 다음 케이스로 이어서 넘기기기
				currentCase = (currentCase + 1) % 4 // 순환 처리
			}
		} else {
			if p2 < 0.8 {
				// 이전 케이스로 넘어왔는데 DEL인 경우
				// 높은 확률로 이전전 케이스로 이어서 넘기기기
				currentCase = (currentCase + 3) % 4 // 순환 처리
			}
		}
	}

	// 해당 케이스의 범위에서 랜덤 값 생성
	switch currentCase {
	case 0: // 0~31 스페셜 문자
		b = uint8(testing.RandRange(ABSOLUTE_MIN, SPECIAL_MAX))
	case 1: // 32~126 아스키
		b = uint8(testing.RandRange(LOW_ASCII_MIN, LOW_ASCII_MAX))
	case 2:
		// DEL문자
		b = DEL
	case 3: // 128~255 하이 아스키
		b = uint8(testing.RandRange(HIGH_ASCII_MIN, HIGH_ASCII_MAX))

	}

	return byte(b)
}

const (
	ABSOLUTE_MIN = 0x00
	SPECIAL_MAX  = 0x1F

	LOW_ASCII_MIN = 0x20
	LOW_ASCII_MAX = 0x7E
	DEL           = 0x7F

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
	p := testing.GetSingleRand().Float64() // 0.0 <= p < 1.0

	// 단지, del확률 처리용으로 만든 변수. 신경 쓸 필요 없음.
	next_flag := testing.GenerateRandomBool(0.5)

	var r int64
	// 현재 케이스를 판별
	currentCase := determineCase(seed_rune)

	// 확률에 따른 케이스 이동
	// 점진적인
	if currentCase == 2 {
		// DEL케이스인 경우 확률 조정
		switch {
		case p < 0.45: // 이전 케이스
			currentCase = (currentCase + 6) % 7 // 순환 처리
			break
		case p < 0.55: // 본인 케이스 유지 확률은 10%
			break
		// 유지
		case p <= 1.0: // 다음 케이스
			currentCase = (currentCase + 1) % 7 // 순환 처리
		}
	} else {
		switch {
		case p < 0.15: // 이전 케이스
			currentCase = (currentCase + 5) % 7 // 순환 처리
			next_flag = false
			break
		case p < 0.35:
			currentCase = (currentCase + 6) % 7 // 순환 처리
			next_flag = false
			break
		case p < 0.65: // 본인 케이스
			break
			// 유지
		case p < 0.85:
			currentCase = (currentCase + 1) % 7 // 순환 처리
			next_flag = true
			break
		case p <= 1.0: // 다음 케이스
			next_flag = true
			currentCase = (currentCase + 2) % 7 // 순환 처리
		}
	}

	if currentCase == 2 {
		p2 := testing.GetSingleRand().Float64()
		// DEL케이스인 경우 확률 조정
		if next_flag {
			if p2 < 0.8 {
				// 다음 케이스로 넘어왔는데 DEL인 경우
				// 높은 확률로 다음 케이스로 이어서 넘기기기
				currentCase = (currentCase + 1) % 7 // 순환 처리
			}
		} else {
			if p2 < 0.8 {
				// 이전 케이스로 넘어왔는데 DEL인 경우
				// 높은 확률로 이전전 케이스로 이어서 넘기기기
				currentCase = (currentCase + 6) % 7 // 순환 처리
			}
		}

	}

	// 해당 케이스의 범위에서 랜덤 값 생성
	switch currentCase {
	case 0: // 0~31 스페셜 문자
		r = testing.RandRange(ABSOLUTE_MIN, SPECIAL_MAX)
	case 1: // 32~126 아스키
		r = testing.RandRange(LOW_ASCII_MIN, LOW_ASCII_MAX)
	case 2:
		// DEL문자
		r = DEL
	case 3: // 128~255 하이 아스키
		r = testing.RandRange(HIGH_ASCII_MIN, HIGH_ASCII_MAX)
	case 4: // 그 외 프린트 가능한 유니코드
		r = testing.RandRange(PRINT_UNICODE_MIN, PRINT_UNICODE_MAX)
	case 5: // 그 외의 프린트 불가능한 유니코드
		r = testing.RandRange(BOUNDARY_UNICODE_MIN, BOUNDARY_UNICODE_MAX)
	case 6:
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
	// DEL 문자 케이스
	case seed_rune == DEL:
		return 2
	// 128 ~256의 높은 아스키 케이스
	case seed_rune >= HIGH_ASCII_MIN && seed_rune <= HIGH_ASCII_MAX:
		return 3
	// 프린트 가능한 유니코드 케이스
	case seed_rune >= PRINT_UNICODE_MIN && seed_rune <= PRINT_UNICODE_MAX:
		return 4
	// 프린트 불가능한 유니코드 케이스
	case seed_rune >= BOUNDARY_UNICODE_MIN && seed_rune <= BOUNDARY_UNICODE_MAX:
		return 5
	// 유니코드가 아닌 케이스
	default:
		return 6
	}
}

// 랜덤 변형 함수 for int types
func randomInt_from(i interface{}) interface{} {
	p := testing.GetSingleRand().Float64() // 0.0 ~ 1.0 사이의 랜덤 확률 값

	switch v := i.(type) {
	case int:
		var int_std interface{}

		switch {
		case p < 0.15:
			min := int64(v) - 100
			max := int64(v) + 100
			if min > max {
				min, max = max, min
			}
			int_std = int(testing.RandRange(min, max))
		case p < 0.3:
			min := int64(v) - 1000
			max := int64(v) + 1000
			if min > max {
				min, max = max, min
			}
			int_std = int(testing.RandRange(min, max))
		case p < 0.45:
			min := int64(v) - 10000
			max := int64(v) + 10000
			if min > max {
				min, max = max, min
			}
			int_std = int(testing.RandRange(min, max))
		case p < 0.60:
			min := int64(v) - 100000
			max := int64(v) + 100000
			if min > max {
				min, max = max, min
			}
			int_std = int(testing.RandRange(min, max))
		default: // 나머지 확률: 완전 랜덤 값
			int_std = testing.GetSingleRand().Int()
		}
		return int_std

	case int8:
		var int_8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) - 5
			max := int64(v) + 5
			if min > max {
				min, max = max, min
			}
			int_8 = int8(testing.RandRange(min, max))
		case p < 0.5:
			min := int64(v) - 30
			max := int64(v) + 30
			if min > max {
				min, max = max, min
			}
			int_8 = int8(testing.RandRange(min, max))
		default: // 나머지 확률: 완전 랜덤 값
			int_8 = int8(testing.RandRange(-128, 127))
		}
		return int_8

	case int16:
		var int_16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) - 100
			max := int64(v) + 100
			if min > max {
				min, max = max, min
			}
			int_16 = int16(testing.RandRange(min, max))
		case p < 0.5:
			min := int64(v) - 1000
			max := int64(v) + 1000
			if min > max {
				min, max = max, min
			}
			int_16 = int16(testing.RandRange(min, max))
		default: // 나머지 확률: 완전 랜덤 값
			int_16 = int16(testing.RandRange(-32768, 32767))
		}
		return int_16

	case int32:
		var int_32 interface{}
		switch {
		case p < 0.2:
			min := int64(v) - 100
			max := int64(v) + 100
			if min > max {
				min, max = max, min
			}
			int_32 = int32(testing.RandRange(min, max))
		case p < 0.4:
			min := int64(v) - 1000
			max := int64(v) + 1000
			if min > max {
				min, max = max, min
			}
			int_32 = int32(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) - 10000
			max := int64(v) + 10000
			if min > max {
				min, max = max, min
			}
			int_32 = int32(testing.RandRange(min, max))
		default: // 나머지 확률: 완전 랜덤 값
			int_32 = testing.GetSingleRand().Int32()
		}
		return int_32

	case int64:
		var int_64 interface{}
		switch {
		case p < 0.15:
			min := v - 100
			max := v + 100
			if min > max {
				min, max = max, min
			}
			int_64 = testing.RandRange(min, max)
		case p < 0.3:
			min := v - 1000
			max := v + 1000
			if min > max {
				min, max = max, min
			}
			int_64 = testing.RandRange(min, max)
		case p < 0.45:
			min := v - 10000
			max := v + 10000
			if min > max {
				min, max = max, min
			}
			int_64 = testing.RandRange(min, max)
		case p < 0.60:
			min := v - 100000
			max := v + 100000
			if min > max {
				min, max = max, min
			}
			int_64 = testing.RandRange(min, max)
		default: // 나머지 확률: 완전 랜덤 값
			int_64 = testing.GetSingleRand().Int64()
		}
		return int_64

	default:
		panic("it's not supported int type")
	}
}

// 랜덤 변형 함수 for uint types
func randomUint_from(u interface{}) interface{} {
	p := testing.GetSingleRand().Float64() // 0.0 ~ 1.0 사이의 랜덤 확률 값

	switch v := u.(type) {
	case uint:
		var uint_std interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 2
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			uint_std = uint(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 10
			max := int64(v) * 10
			if min > max {
				min, max = max, min
			}
			uint_std = uint(testing.RandRange(min, max))
		default:
			uint_std = uint(testing.GetSingleRand().Uint64())
		}
		return uint_std

	case uint8:
		var uint_8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 2
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			uint_8 = uint8(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 10
			max := int64(v) * 10
			if min > max {
				min, max = max, min
			}
			uint_8 = uint8(testing.RandRange(min, max))
		default:
			uint_8 = uint8(testing.RandRange(0, 255)) // uint8 범위: 0 ~ 255
		}
		return uint_8

	case uint16:
		var uint_16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 2
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			uint_16 = uint16(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 10
			max := int64(v) * 10
			if min > max {
				min, max = max, min
			}
			uint_16 = uint16(testing.RandRange(min, max))
		default:
			uint_16 = uint16(testing.RandRange(0, 65535)) // uint16 범위: 0 ~ 65535
		}
		return uint_16

	case uint32:
		var uint_32 interface{}
		switch {
		case p < 0.2:
			min := int64(v) / 2
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			uint_32 = uint32(testing.RandRange(min, max))
		case p < 0.4:
			min := int64(v) / 8
			max := int64(v) * 8
			if min > max {
				min, max = max, min
			}
			uint_32 = uint32(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 16
			max := int64(v) * 16
			if min > max {
				min, max = max, min
			}
			uint_32 = uint32(testing.RandRange(min, max))
		default:
			uint_32 = uint32(testing.RandRange(0, int64(^uint32(0))))
		}
		return uint_32

	case uint64:
		var uint_64 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 2
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			uint_64 = uint64(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 8
			max := int64(v) * 8
			if min > max {
				min, max = max, min
			}
			uint_64 = uint64(testing.RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 32
			max := int64(v) * 32
			if min > max {
				min, max = max, min
			}
			uint_64 = uint32(testing.RandRange(min, max))
		default:
			uint_64 = testing.RandRange(0, int64(uint64(0))) // uint64 범위: 0 ~ 18446744073709551615
		}
		return uint_64

	default:
		panic("it's not a supported uint type")
	}
}

func randomFloat_from(f interface{}) interface{} {
	switch v := f.(type) {
	case float32:
		var f_32 float32
		f_32 = randFloat32_from(float32(v))
		return float32(f_32)
	case float64:
		var f_64 float64
		f_64 = randFloat64_from(float64(v))
		return float64(f_64)
	default:
		panic("argument is not float 32 or float4")
	}
}

func randFloat32_from(f float32) float32 {
	return 0.0000000
}

func randFloat64_from(f float64) float64 {
	return 0.0000000
}

func randomBool() bool {
	return testing.UniformRandomBool(0.5)
}

func linear_crossover(a interface{}, b interface{}) (interface{}, interface{}) {
	switch v1 := a.(type) {
	case int:
		v2, ok := b.(int)
		if !ok {
			panic("type not equal")
		}
		new_v1 := int(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := int(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case int8:
		v2, ok := b.(int8)
		if !ok {
			panic("type not equal")
		}
		new_v1 := int8(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := int8(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case int16:
		v2, ok := b.(int16)
		if !ok {
			panic("type not equal")
		}
		new_v1 := int16(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := int16(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case int32:
		v2, ok := b.(int32)
		if !ok {
			panic("type not equal")
		}
		new_v1 := int32(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := int32(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case int64:
		v2, ok := b.(int64)
		if !ok {
			panic("type not equal")
		}
		new_v1 := int64(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := int64(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case uint:
		v2, ok := b.(uint)
		if !ok {
			panic("type not equal")
		}
		new_v1 := uint(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := uint(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case uint8:
		v2, ok := b.(uint8)
		if !ok {
			panic("type not equal")
		}
		new_v1 := uint8(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := uint8(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case uint16:
		v2, ok := b.(uint16)
		if !ok {
			panic("type not equal")
		}
		new_v1 := uint16(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := uint16(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case uint32:
		v2, ok := b.(uint32)
		if !ok {
			panic("type not equal")
		}
		new_v1 := uint32(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := uint32(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case uint64:
		v2, ok := b.(uint64)
		if !ok {
			panic("type not equal")
		}
		new_v1 := uint64(math.Round(0.7*float64(v1) + 0.3*float64(v2)))
		new_v2 := uint64(math.Round(0.3*float64(v1) + 0.7*float64(v2)))
		return new_v1, new_v2
	case float32:
		v2, ok := b.(float32)
		if !ok {
			panic("type not equal")
		}
		new_v1 := float32(0.7*float64(v1) + 0.3*float64(v2))
		new_v2 := float32(0.3*float64(v1) + 0.7*float64(v2))
		return new_v1, new_v2
	case float64:
		v2, ok := b.(float64)
		if !ok {
			panic("type not equal")
		}
		new_v1 := float64(0.3*float64(v1) + 0.7*float64(v2))
		new_v2 := float64(0.3*float64(v1) + 0.7*float64(v2))
		return new_v1, new_v2
	default:
		panic("type can't be linear crossed")
	}
}
