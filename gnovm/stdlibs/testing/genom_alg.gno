package testing

import (
	"bytes"
	"errors"
	"math"
	"strconv"
	"time"
	"unicode/utf8"
)

func Mutate(seed Seed, mut_range int) Seed {
	if len(seed.Content) == 0 {
		panic("mutate logic error: content's len==0")
	}
	index := 0
	if len(seed.Content) > 1 {
		index = int(RandRange(0, int64(len(seed.Content))))
	}

	selected := seed.Content[index]
	switch v := selected.(type) {
	case int, int8, int16, int32, int64:
		for i := 0; i < mut_range; i++ {
			seed.Content[index] = randomInt_from(v)
		}
	case uint, uint8, uint16, uint32, uint64:
		for i := 0; i < mut_range; i++ {
			seed.Content[index] = randomUint_from(v)
		}
	case float32, float64:
		for i := 0; i < mut_range; i++ {
			seed.Content[index] = randomFloat_from(v)
		}
	case bool:
		seed.Content[index] = randomBool()
	// Cancellation due to value set issue
	// case string:
	// 	runes := []rune(v)
	// 	if len(runes) > 0 {
	// 		runeIndex := RandRange(0, int64(len(runes)))
	// 		runes[runeIndex] = randomRune_from(runes[runeIndex])
	// 	}
	// 	var new_str string = string(runes)
	// 	seed.Content[index] = new_str
	case string:
		bytes := []byte(v)
		if len(bytes) > 0 {
			for i := 0; i < mut_range; i++ {
				byteIndex := RandRange(0, int64(len(bytes)))
				bytes[byteIndex] = randomByte_from(bytes[byteIndex])
			}
		}
		var new_str string = string(bytes)
		seed.Content[index] = new_str

	case []byte:
		bytes := []byte(v)
		if len(bytes) > 0 {
			for i := 0; i < mut_range; i++ {
				byteIndex := RandRange(0, int64(len(bytes)))
				bytes[byteIndex] = randomByte_from(bytes[byteIndex])
			}
		}
		var new_byt []byte = []byte(bytes)
		seed.Content[index] = new_byt
	default:
		panic("not supported type")
	}

	return seed
}

func InsertDelete(seed Seed, p float64, mut_range int, string_byte_candidates []int) Seed {
	if len(string_byte_candidates) == 0 {
		return seed
	}

	index := 0
	if len(string_byte_candidates) > 0 {

		selected_field_idx := RandRange(0, int64(len(string_byte_candidates)))
		index = string_byte_candidates[selected_field_idx]
	}

	selected := seed.Content[index]

	switch v := selected.(type) {
	case []byte:
		bb := []byte(v)
		for i := 0; i < mut_range; i++ {
			l := len(bb)
			// Insert
			if GenerateRandomBool(p) {
				if l < 1 {
					var b byte = ' '
					bb = []byte{randomByte_from(b)}
				} else {

					sample := bb[RandRange(0, int64(l))]
					bt := randomByte_from(sample)

					pos := RandRange(0, int64(l))

					bb = append(bb, 0)

					copy(bb[pos+1:], bb[pos:])

					bb[pos] = bt
				}
			} else {
				// Del
				if l == 0 {
					return seed
				}
				pos := RandRange(0, int64(l))
				bb = append(bb[:pos], bb[pos+1:]...)
			}

		}
		var new_byt []byte = bb
		seed.Content[index] = new_byt
	case string:
		bb := []byte(v)
		for i := 0; i < mut_range; i++ {
			l := len(bb)
			// Insert
			if GenerateRandomBool(p) {
				if l < 1 {
					var b byte = ' '
					bb = []byte{randomByte_from(b)}
				} else {

					sample := bb[RandRange(0, int64(l))]
					bt := randomByte_from(sample)

					pos := RandRange(0, int64(l))

					bb = append(bb, 0)

					copy(bb[pos+1:], bb[pos:])

					bb[pos] = bt
				}
			} else {
				if l != 0 {
					pos := RandRange(0, int64(l))
					bb = append(bb[:pos], bb[pos+1:]...)
				}
			}

		}
		var new_str string = string(bb)
		seed.Content[index] = new_str
	default:
		println("maybe some nil string error:", v)
		panic("internal logic error")
	}
	return seed
}

// I deleted the existing fit, fitness.
// As I tried to increase the speed by integrating into AFL, I decided that it was faster to manage it with just queue, stack, and unique linked list.
// (Something gets uncomfortable if you follow the afl logic I've seen and maintain that fitness management.)
// Fitness and selection logic are replaced.

// Modified existing mating logic.
// I adjusted the number according to gen to solve the sticking problem.
// Changed to multi-intersection logic.
func TwoPointCrossover(parent1, parent2 Seed, seedCount *uint) (Seed, Seed) {
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
			content1[i], content2[i] = factorization_crossover(v1, content2[i])
		case bool:
			content1[i] = v1
			content2[i] = content2[i]

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
			if minLen == 0 {
				maxLen := p1Len
				m := 1
				if p2Len > p1Len {
					m = 2
					maxLen = p2Len
				}
				if maxLen < 1 {
					s := ' '
					bb := byte(s)
					content1[i] = []byte([]byte{randomByte_from(bb)})
					content2[i] = []byte([]byte{randomByte_from(bb)})
					continue
				} else {
					if m == 1 {
						content2[i] = content1[i]
					} else {
						content1[i] = content2[i]
					}
					continue
				}
			}

			point1 := RandRange(0, int64(minLen))
			point2 := RandRange(0, int64(minLen))

			if point1 > point2 {
				point1, point2 = point2, point1
			}

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
		case string:
			// t1 := tokenizeString(v1)
			// v2, ok := parent2.Content[i].(string)
			// if !ok {
			// 	panic("type not equal")
			// }
			// t2 := tokenizeString(v2)

			// c1, c2 := twoPointCrossoverTokens(t1, t2)
			// result_str1 := rebuildString(c1)
			// result_str2 := rebuildString(c2)

			byt1 := v1
			byt2, ok := parent2.Content[i].(string)
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
			if minLen < 1 {
				maxLen := p1Len
				m := 1
				if p2Len > p1Len {
					m = 2
					maxLen = p2Len
				}
				if maxLen < 1 {
					s := ' '
					bb := byte(s)
					content1[i] = string(randomByte_from(bb))
					content2[i] = string(randomByte_from(bb))
					continue
				} else {
					if m == 1 {
						content2[i] = content1[i]
					} else {
						content1[i] = content2[i]
					}
					continue
				}
			}

			point1 := RandRange(0, int64(minLen))
			point2 := RandRange(0, int64(minLen))

			if point1 > point2 {
				point1, point2 = point2, point1
			}

			crossed_byt1 := append([]byte{}, p1Bytes[:point1]...)          // 부모1의 첫 구간
			crossed_byt1 = append(crossed_byt1, p2Bytes[point1:point2]...) // 부모2의 중간 구간
			crossed_byt1 = append(crossed_byt1, p1Bytes[point2:]...)       // 부모1의 마지막 구간

			crossed_byt2 := append([]byte{}, p2Bytes[:point1]...)          // 부모2의 첫 구간
			crossed_byt2 = append(crossed_byt2, p1Bytes[point1:point2]...) // 부모1의 중간 구간
			crossed_byt2 = append(crossed_byt2, p2Bytes[point2:]...)       // 부모2의 마지막 구간

			result_str1 := string(crossed_byt1)
			result_str2 := string(crossed_byt2)
			content1[i] = result_str1
			content2[i] = result_str2

		default:
			panic("not supported type")

		}
	}

	*seedCount++
	updatedIdv1 := Seed{
		Gen: parent1.Gen + 1, IsCoordinated: false, Content: content1, Result: nil,
		Pid: parent1.Id, Id: *seedCount,
	}
	*seedCount++
	updatedIdv2 := Seed{
		Gen: parent2.Gen + 1, IsCoordinated: false, Content: content2, Result: nil,
		Pid: parent1.Id, Id: *seedCount,
	}

	return updatedIdv1, updatedIdv2
}

// Returns the parents and returns the child.
// I fluidized the number and degree of mating along Gen
// Seeds received as parameters are invariant referenced within the function.
func Evolve(seeds []Seed, seedCount *uint, string_byte_candidates []int) []Seed {
	p1 := seeds[0]
	Mutation_Range := [10]int{5, 4, 3, 3, 3, 2, 2, 2, 2, 2}
	var mutation_range int
	if int(p1.Gen) > len(Mutation_Range) {
		mutation_range = 1
	} else {
		mutation_range = Mutation_Range[p1.Gen-1]
	}
	Evlove_Count := [10]int{1920, 1440, 1200, 1024, 960, 720, 600, 480, 360, 240}
	var evolve_count int
	if int(p1.Gen) > len(Evlove_Count) {
		evolve_count = 240
	} else {
		evolve_count = Evlove_Count[p1.Gen-1]
	}

	loop_count := evolve_count / 2

	new_generation := []Seed{}

	if len(seeds) == 1 {

		for i := 0; i < int(evolve_count); i++ {
			new_content := make([]interface{}, len(seeds[0].Content))
			for i, v := range seeds[0].Content {
				new_content[i] = v
			}
			*seedCount++
			new_ind := Seed{
				Gen:           seeds[0].Gen + 1,
				IsCoordinated: false,
				Content:       new_content,
				Result:        nil,
				Pid:           seeds[0].Id,
				Id:            *seedCount,
			}

			if UniformRandomBool(0.6) {
				new_ind = Mutate(new_ind, mutation_range)
			}
			if UniformRandomBool(0.4) {
				new_ind = InsertDelete(new_ind, 0.5, mutation_range, string_byte_candidates)
			}

			new_ind.Gen = seeds[0].Gen + 1
			new_generation = append(new_generation, new_ind)

		}
		return new_generation
	}

	if len(seeds) > 3 {
		panic("not covered len")
	}

	p2 := seeds[1]

	for i := 0; i < int(loop_count); i++ {
		c1, c2 := TwoPointCrossover(p1, p2, seedCount)
		new_generation = append(new_generation, c1)
		new_generation = append(new_generation, c2)

	}

	for i := range new_generation {

		if UniformRandomBool(0.4) {
			new_generation[i] = Mutate(new_generation[i], mutation_range) // 원본 데이터 수정
		}
		if UniformRandomBool(0.3) {
			new_generation[i] = InsertDelete(new_generation[i], 0.5, mutation_range, string_byte_candidates)
		}

	}

	return new_generation
}

func randomByte_from(seed_byte byte) byte {
	p := GetSingleRand().Float64() // 0.0 <= p < 1.0

	next_flag := GenerateRandomBool(0.5)
	var b uint8

	currentCase := determineCase(rune(seed_byte))

	if currentCase == 2 {
		switch {
		case p < 0.45:
			currentCase = (currentCase + 3) % 4
			break
		case p < 0.55:
			break

		case p <= 1.0:
			currentCase = (currentCase + 1) % 4
		}
	} else {
		switch {
		case p < 0.1:
			currentCase = (currentCase + 2) % 4
			next_flag = false
			break
		case p >= 0.1 && p < 0.3:
			currentCase = (currentCase + 3) % 4
			next_flag = false
			break
		case p >= 0.3 && p < 0.7:
			break

		case p >= 0.7 && p < 0.9:
			next_flag = true
			currentCase = (currentCase + 1) % 4
		case p >= 0. && p <= 1.0:
			next_flag = true
			currentCase = (currentCase + 2) % 4
		}
	}

	if currentCase == 2 {
		p2 := GetSingleRand().Float64()
		if next_flag {
			if p2 < 0.8 {
				currentCase = (currentCase + 1) % 4
			}
		} else {
			if p2 < 0.8 {
				currentCase = (currentCase + 3) % 4
			}
		}
	}

	switch currentCase {
	case 0:
		b = uint8(RandRange(ABSOLUTE_MIN, SPECIAL_MAX+1))
	case 1:
		b = uint8(RandRange(LOW_ASCII_MIN, LOW_ASCII_MAX+1))
	case 2:

		b = DEL
	case 3:
		b = uint8(RandRange(HIGH_ASCII_MIN, HIGH_ASCII_MAX+1))

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

func determineCase(seed_rune rune) int {
	switch {
	case seed_rune >= ABSOLUTE_MIN && seed_rune <= SPECIAL_MAX:
		return 0
	case seed_rune >= LOW_ASCII_MIN && seed_rune <= LOW_ASCII_MAX:
		return 1
	case seed_rune == DEL:
		return 2
	case seed_rune >= HIGH_ASCII_MIN && seed_rune <= HIGH_ASCII_MAX:
		return 3
	case seed_rune >= PRINT_UNICODE_MIN && seed_rune <= PRINT_UNICODE_MAX:
		return 4
	case seed_rune >= BOUNDARY_UNICODE_MIN && seed_rune <= BOUNDARY_UNICODE_MAX:
		return 5
	default:
		return 6
	}
}

func randomInt_from(i interface{}) interface{} {
	p := GetSingleRand().Float64()

	switch v := i.(type) {
	case int:
		var int_std interface{}
		if v == 0 {
			return int(RandInt64())
		}
		switch {
		case p < 0.15:
			min := int64(v) * (-2)
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			int_std = int(RandRange(min, max))
		case p < 0.3:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			int_std = int(RandRange(min, max))
		case p < 0.45:
			min := int64(v) * (-8)
			max := int64(v) * (8)
			if min > max {
				min, max = max, min
			}
			int_std = int(RandRange(min, max))
		case p < 0.60:
			min := int64(v) * (-16)
			max := int64(v) * (16)
			if min > max {
				min, max = max, min
			}
			int_std = int(RandRange(min, max))
		default:
			int_std = GetSingleRand().Int()
		}
		return int_std

	case int8:
		if v == 0 {
			return int8(RandInt64())
		}
		var int_8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			int_8 = int8(RandRange(min, max))
		case p < 0.5:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			int_8 = int8(RandRange(min, max))
		default:
			int_8 = int8(RandRange(-128, 128))
		}
		return int_8

	case int16:
		if v == 0 {
			return int16(RandInt64())
		}
		var int_16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			int_16 = int16(RandRange(min, max))
		case p < 0.5:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			int_16 = int16(RandRange(min, max))
		default:
			int_16 = int16(RandRange(-32768, 32768))
		}
		return int_16

	case int32:
		if v == 0 {
			return int32(RandInt64())
		}
		var int_32 interface{}
		switch {
		case p < 0.2:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			int_32 = int32(RandRange(min, max))
		case p < 0.4:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			int_32 = int32(RandRange(min, max))
		case p < 0.6:
			min := int64(v) * (-8)
			max := int64(v) * (8)
			if min > max {
				min, max = max, min
			}
			int_32 = int32(RandRange(min, max))
		default:
			int_32 = GetSingleRand().Int32()
		}
		return int_32

	case int64:
		if v == 0 {
			return RandInt64()
		}
		var int_64 interface{}
		switch {
		case p < 0.15:
			min := v * (-2)
			max := v * (2)
			if min > max {
				min, max = max, min
			}
			int_64 = RandRange(min, max)
		case p < 0.3:
			min := v * (-4)
			max := v * (4)
			if min > max {
				min, max = max, min
			}
			int_64 = RandRange(min, max)
		case p < 0.45:
			min := v * (-8)
			max := v * (8)
			if min > max {
				min, max = max, min
			}
			int_64 = RandRange(min, max)
		case p < 0.60:
			min := v * (-16)
			max := v * (16)
			if min > max {
				min, max = max, min
			}
			int_64 = RandRange(min, max)
		default:
			int_64 = GetSingleRand().Int64()
		}
		return int_64

	default:
		panic("it's not supported int type")
	}
}

func randomUint_from(u interface{}) interface{} {
	p := GetSingleRand().Float64()

	switch v := u.(type) {
	case uint:
		if v == 0 {
			return uint(RandUint64())
		}
		var uint_std interface{}
		switch {
		case p < 0.3:
			min := v / 256
			max := v * 2
			uint_std = uint(UintRandRange(uint64(min), uint64(max)))
		case p < 0.6:
			min := v / 9096
			max := v * 8
			uint_std = uint(UintRandRange(uint64(min), uint64(max)))
		default:
			uint_std = uint(GetSingleRand().Uint64())
		}
		return uint_std

	case uint8:
		if v == 0 {
			return uint8(RandUint64())
		}
		var uint_8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 8
			max := int64(v) * 2
			uint_8 = uint8(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 32
			max := int64(v) * 8
			uint_8 = uint8(RandRange(min, max))
		default:
			uint_8 = uint8(RandRange(0, 256))
		}
		return uint_8

	case uint16:
		if v == 0 {
			return uint16(RandUint64())
		}
		var uint_16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 256
			max := int64(v) * 2
			uint_16 = uint16(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 9096
			max := int64(v) * 8
			uint_16 = uint16(RandRange(min, max))
		default:
			uint_16 = uint16(RandRange(0, 65536))
		}
		return uint_16

	case uint32:
		if v == 0 {
			return uint32(RandUint64())
		}
		var uint_32 interface{}
		switch {
		case p < 0.2:
			min := int64(v) / 256
			max := int64(v) * 2
			uint_32 = uint32(RandRange(min, max))
		case p < 0.4:
			min := int64(v) / 9096
			max := int64(v) * 8
			uint_32 = uint32(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / (9096 * 9096)
			max := int64(v) * 16
			uint_32 = uint32(RandRange(min, max))
		default:
			uint_32 = uint32(RandRange(0, int64(^uint32(0))))
		}
		return uint_32

	case uint64:
		if v == 0 {
			return RandUint64()
		}
		var uint_64 interface{}
		switch {
		case p < 0.2:
			min := v / 256
			max := v * 2
			uint_64 = UintRandRange(min, max)
		case p < 0.5:
			min := v / 9096
			max := v * 8
			uint_64 = UintRandRange(min, max)
		case p < 0.5:
			min := v / (9096 * 9096)
			max := v * 16
			uint_64 = UintRandRange(min, max)
		default:
			uint_64 = RandUint64()
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
	return randomFloat32(f)
}

func randFloat64_from(f float64) float64 {
	return randomFloat64(f)
}

func randomBool() bool {
	return UniformRandomBool(0.5)
}

func factorization_crossover(a interface{}, b interface{}) (interface{}, interface{}) {
	switch v1 := a.(type) {
	case int:
		v2, ok := b.(int)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var new_v1 int
		var new_v2 int
		divisor := int(RandRange(1, int64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2
	case int8:
		v2, ok := b.(int8)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var new_v1 int8
		var new_v2 int8
		divisor := int8(RandRange(1, int64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2
	case int16:
		v2, ok := b.(int16)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var new_v1 int16
		var new_v2 int16
		divisor := int16(RandRange(1, int64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}

		return new_v1, new_v2
	case int32:
		v2, ok := b.(int32)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var new_v1 int32
		var new_v2 int32
		divisor := int32(RandRange(1, int64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2

	case int64:
		v2, ok := b.(int64)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var new_v1 int64
		var new_v2 int64
		divisor := RandRange(1, int64(min)/2)
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2

	case uint:
		v2, ok := b.(uint)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var new_v1 uint
		var new_v2 uint
		divisor := uint(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2
	case uint8:
		v2, ok := b.(uint8)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var new_v1 uint8
		var new_v2 uint8
		divisor := uint8(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2

	case uint16:
		v2, ok := b.(uint16)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var new_v1 uint16
		var new_v2 uint16
		divisor := uint16(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2

	case uint32:
		v2, ok := b.(uint32)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var new_v1 uint32
		var new_v2 uint32
		divisor := uint32(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
		return new_v1, new_v2

	case uint64:
		v2, ok := b.(uint64)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var new_v1 uint64
		var new_v2 uint64
		divisor := uint64(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			new_v1 = v1 / divisor
			new_v2 = v2 * divisor
		} else {
			new_v1 = v1 * divisor
			new_v2 = v2 / divisor
		}
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
		panic("type can't be  factorization_crossovered.")
	}
}
