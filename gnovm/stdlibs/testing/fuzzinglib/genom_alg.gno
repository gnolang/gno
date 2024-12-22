package fuzzinglib

import (
	"testing"
)

type Seed struct {
	Gen     uint
	IsUsed  bool
	Content string
	Result  interface{}
}

func randomASCIIChar() rune {
	r := int(testing.RandRange(32, 96))
	return rune(r)
}

// func main() {
// 	seed1 := Seed{Gen: 3, IsUsed: false, Content: "apple", Result: nil}

// 	seed2 := Seed{Gen: 3, IsUsed: false, Content: "banana", Result: nil}

// 	seeds := evolve([]Seed{seed1, seed2})

// 	o_seeds := evolve([]Seed{seed1})

// 	println(seeds)
// 	println(len(seeds))
// 	println(o_seeds)
// 	println(len(o_seeds))
// }

// Seed의 Content를 랜덤으로 변형
func Mutate(seed Seed) Seed {
	runes := []rune(seed.Content)
	if len(runes) == 0 {
		return seed
	}

	index := testing.RandRange(0, len(runes)-1)
	runes[index] = randomASCIIChar()
	seed.Content = string(runes)

	return seed
}

// p확률로 랜덤 삽입. 아님 랜덤 삭제
func InsertDelete(seed Seed, p float64) Seed {
	s := seed.Content
	rr := []rune(s)
	l := len(rr)

	// Insert
	if testing.GenerateRandomBool(p) {
		pos := testing.RandRange(0, l-1)
		rr = append(rr, 0)

		copy(rr[pos+1:], rr[pos:])

		char := randomASCIIChar()
		rr[pos] = char
	} else {
		if l == 0 {
			return seed
		}

		pos := testing.RandRange(0, l-1)
		rr = append(rr[:pos], rr[pos+1:]...)
	}
	seed.Content = string(rr)

	return seed
}

// 기존의 적합도, 피트니스를 삭제했습니다.
// AFl에 통합해서 속도를 높이려다 보니 그냥 큐, 스택, 링크드 리스트로 관리하는게 빠르다 판단했습니다.
// (제가 본 afl로직을 따라가면서 해당 피트니스 관리를 유지하게되면 뭔가 불편해집니다.)
// 피트니스, 선택 로직은 치환된 것이라 보면 되겠습니다.

// 기존의 교배 로직을 수정했습니다.
// 고착화 문제 해결을 위해 gen에 따른 수를 조정했습니다.
// 다중 교차점 로직으로 바꿨습니다.
func TwoPointCrossover(parent1, parent2 Seed) (Seed, Seed) {
	p1Runes := []rune(parent1.Content)
	p2Runes := []rune(parent2.Content)

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
	child1 := append([]rune{}, p1Runes[:point1]...)    // 부모1의 첫 구간
	child1 = append(child1, p2Runes[point1:point2]...) // 부모2의 중간 구간
	child1 = append(child1, p1Runes[point2:]...)       // 부모1의 마지막 구간

	child2 := append([]rune{}, p2Runes[:point1]...)    // 부모2의 첫 구간
	child2 = append(child2, p1Runes[point1:point2]...) // 부모1의 중간 구간
	child2 = append(child2, p2Runes[point2:]...)       // 부모2의 마지막 구간

	// 새로운 자식 생성
	updatedIdv1 := Seed{Gen: parent1.Gen + 1, IsUsed: false, Content: string(child1), Result: nil}
	updatedIdv2 := Seed{Gen: parent2.Gen + 1, IsUsed: false, Content: string(child2), Result: nil}

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
			var new_ind Seed
			if testing.UniformRandomBool(0.6) {
				new_ind = Mutate(seeds[0])
			} else {
				new_ind = InsertDelete(seeds[0], 0.3)
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
		new_generation = append(new_generation, c1, c2)
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
