package testing

import (
	"math"
	"time"
)

// Returns the parents and returns the child.
// fluidized the number and degree of mutation along Gen
// Seeds received as parameters are invariant referenced within the function.
func evolve(seeds []seed, seedCount *uint, stringByteCandidates []int) []seed {
	p1 := seeds[0]
	MutationStrength := [10]int{5, 4, 3, 3, 2, 2, 2, 2, 2, 2}
	var mutationStrength int
	if int(p1.gen) > len(MutationStrength) {
		mutationStrength = 1
	} else {
		mutationStrength = MutationStrength[p1.gen-1]
	}
	CreatingCount := [10]int{1920, 1440, 1200, 1000, 960, 720, 600, 480, 360, 240}
	var creatingCount int
	if int(p1.gen) > len(CreatingCount) {
		creatingCount = 240
	} else {
		creatingCount = CreatingCount[p1.gen-1]
	}

	loopCount := creatingCount / 2

	newGeneration := []seed{}

	if len(seeds) == 1 {

		for i := 0; i < int(creatingCount); i++ {
			newContent := make([]interface{}, len(seeds[0].content))
			for i, v := range seeds[0].content {
				newContent[i] = v
			}
			*seedCount++
			newInd := seed{
				gen:           seeds[0].gen + 1,
				isCoordinated: false,
				content:       newContent,
				result:        nil,
				pid:           seeds[0].id,
				id:            *seedCount,
			}

			randSeed := i % 10
			switch randSeed {
			case 0, 1, 2, 3:
				// Variations that are flexible with 50% probability
				newInd = mutate(newInd, mutationStrength)
			case 4, 5, 6:
				newInd = insertDelete(newInd, mutationStrength, stringByteCandidates)
			case 7, 8:
				// Variations of only the degree of uniformity with 10% probability. Maintaining the characteristics of the seed.
				newInd = mutate(newInd, 2)
			case 9:
				// Variate strongly with 10% probability
				newInd = mutate(newInd, mutationStrength)
				newInd = insertDelete(newInd, mutationStrength, stringByteCandidates)
			}

			newInd.gen = seeds[0].gen + 1
			newGeneration = append(newGeneration, newInd)

		}
		return newGeneration
	}

	if len(seeds) > 3 {
		panic("not covered len")
	}

	p2 := seeds[1]

	for i := 0; i < int(loopCount); i++ {
		c1, c2 := twoPointCrossover(p1, p2, seedCount)
		newGeneration = append(newGeneration, c1)
		newGeneration = append(newGeneration, c2)

	}

	for i := range newGeneration {
		randSeed := i % 10
		switch randSeed {
		case 0, 1, 2, 3:
			newGeneration[i] = mutate(newGeneration[i], mutationStrength)
		case 4, 5, 6:
			// do nothing
		case 7, 8:
			newGeneration[i] = insertDelete(newGeneration[i], mutationStrength, stringByteCandidates)
		case 9:
			newGeneration[i] = mutate(newGeneration[i], mutationStrength)
			newGeneration[i] = insertDelete(newGeneration[i], mutationStrength, stringByteCandidates)
		}

	}

	return newGeneration
}

func mutate(seed seed, mutationStrength int) seed {
	index := 0
	if len(seed.content) > 1 {
		index = int(RandRange(0, int64(len(seed.content))))
	}

	selected := seed.content[index]
	switch v := selected.(type) {
	case int, int8, int16, int32, int64:
		for i := 0; i < mutationStrength; i++ {
			seed.content[index] = randomIntFrom(v)
		}
	case uint, uint8, uint16, uint32, uint64:
		for i := 0; i < mutationStrength; i++ {
			seed.content[index] = randomUintFrom(v)
		}
	case float32, float64:
		for i := 0; i < mutationStrength; i++ {
			seed.content[index] = randomFloatFrom(v)
		}
	case bool:
		seed.content[index] = randomBool()
	// String is treated as []byte rather than []rune
	// []rune is not allowed to contain escaped bytes, but string can contain them.
	// https://go.dev/blog/strings
	// so all strings must be treated as []byte in fuzzing.
	// This is the description of the logic that deals with strings as bytes.
	case string:
		bytes := []byte(v)
		if len(bytes) > 0 {
			for i := 0; i < mutationStrength; i++ {
				byteIndex := RandRange(0, int64(len(bytes)))
				bytes[byteIndex] = randomByteFrom(bytes[byteIndex])
			}
		}
		var newStr string = string(bytes)
		seed.content[index] = newStr

	case []byte:
		bytes := []byte(v)
		if len(bytes) > 0 {
			for i := 0; i < mutationStrength; i++ {
				byteIndex := RandRange(0, int64(len(bytes)))
				bytes[byteIndex] = randomByteFrom(bytes[byteIndex])
			}
		}
		var newByt []byte = []byte(bytes)
		seed.content[index] = newByt
	default:
		panic("not supported type")
	}

	return seed
}

func insertDelete(seed seed, mutationStrength int, stringByteCandidates []int) seed {
	if len(stringByteCandidates) == 0 {
		return seed
	}

	index := 0
	if len(stringByteCandidates) > 0 {
		selectedFieldidx := RandRange(0, int64(len(stringByteCandidates)))
		index = stringByteCandidates[selectedFieldidx]
	}

	selected := seed.content[index]

	switch v := selected.(type) {
	case []byte:
		bb := []byte(v)
		for i := 0; i < mutationStrength; i++ {
			bb = addOrRemoveByte(bb)
		}
		var newByte []byte = bb
		seed.content[index] = newByte
	case string:
		bb := []byte(v)
		for i := 0; i < mutationStrength; i++ {
			bb = addOrRemoveByte(bb)
		}
		var newString string = string(bb)
		seed.content[index] = newString
	default:
		panic("internal logic error")
	}
	return seed
}

func addOrRemoveByte(bb []byte) []byte {
	l := len(bb)
	if GenerateRandomBool(0.6) {
		// Insert
		if l == 0 {
			var b byte = ' '
			bb = []byte{randomByteFrom(b)}
		} else {
			sample := bb[RandRange(0, int64(l))]
			bt := randomByteFrom(sample)
			pos := RandRange(0, int64(l))
			bb = append(bb, 0)
			copy(bb[pos+1:], bb[pos:])
			bb[pos] = bt
		}
		return bb
	} else {
		// Del
		if l == 0 {
			b := byte(RandRange(HighAsciiMin, HighAscillMax+1))
			return []byte{b}
		}
		pos := RandRange(0, int64(l))
		bb = append(bb[:pos], bb[pos+1:]...)
		return bb
	}
}

// I deleted the existing fit, fitness.
// As I tried to increase the speed by integrating into AFL, I decided that it was faster to manage it with just queue, stack, and unique linked list.
// (Something gets uncomfortable if follow the afl logic I've seen and maintain that fitness management.)
// Fitness and selection logic are replaced.

// Modified existing crossover logic.
// I adjusted the number according to gen to solve the sticking problem.(a phenomenon in which they become similar as inputs increase)
// Changed to multi-intersection logic.
func twoPointCrossover(parent1, parent2 seed, seedCount *uint) (seed, seed) {
	content1 := make([]interface{}, len(parent1.content))
	for i, v := range parent1.content {
		content1[i] = v
	}
	content2 := make([]interface{}, len(parent2.content))
	for i, v := range parent2.content {
		content2[i] = v
	}

	for i := 0; i < len(parent1.content); i++ {
		switch v1 := content1[i].(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			content1[i], content2[i] = factorizationCrossover(v1, content2[i])
		case bool:
			content1[i] = v1
			content2[i] = content2[i]

		case []byte:
			byt1 := v1
			byt2, ok := parent2.content[i].([]byte)
			if !ok {
				panic("type not equal")
			}
			p1Bytes := []byte(byt1)
			p2Bytes := []byte(byt2)
			resultByt1, resultByt2 := twoPointCrossoverByte(p1Bytes, p2Bytes)
			content1[i] = resultByt1
			content2[i] = resultByt2
		case string:
			byt1 := v1
			byt2, ok := parent2.content[i].(string)
			if !ok {
				panic("type not equal")
			}
			p1Bytes := []byte(byt1)
			p2Bytes := []byte(byt2)
			resultByt1, resultByt2 := twoPointCrossoverByte(p1Bytes, p2Bytes)
			resultStr1 := string(resultByt1)
			resultStr2 := string(resultByt2)
			content1[i] = resultStr1
			content2[i] = resultStr2
		default:
			panic("not supported type")

		}
	}

	*seedCount++
	updatedIdv1 := seed{
		gen: parent1.gen + 1, isCoordinated: false, content: content1, result: nil,
		pid: parent1.id, id: *seedCount,
	}
	*seedCount++
	updatedIdv2 := seed{
		gen: parent2.gen + 1, isCoordinated: false, content: content2, result: nil,
		pid: parent1.id, id: *seedCount,
	}

	return updatedIdv1, updatedIdv2
}

func twoPointCrossoverByte(p1Bytes []byte, p2Bytes []byte) ([]byte, []byte) {
	p1Len := len(p1Bytes)
	p2Len := len(p2Bytes)
	minLen := p1Len
	minIdx := 1
	if p2Len < p1Len {
		minLen = p2Len
		minIdx = 2
	}
	if minLen == 0 {
		maxLen := p1Len
		if minIdx == 1 {
			maxLen = p2Len
		}
		var seedByte byte
		if maxLen == 0 {
			seedByte = byte(' ')
		} else {
			if minIdx == 1 {
				seedByte = p2Bytes[0]
			} else {
				seedByte = p1Bytes[0]
			}
		}
		return []byte([]byte{randomByteFrom(seedByte)}), []byte([]byte{randomByteFrom(seedByte)})
	}

	point1 := RandRange(0, int64(minLen))
	point2 := RandRange(0, int64(minLen))

	if point1 > point2 {
		point1, point2 = point2, point1
	}

	crossedByt1 := append([]byte{}, p1Bytes[:point1]...)
	crossedByt1 = append(crossedByt1, p2Bytes[point1:point2]...)
	crossedByt1 = append(crossedByt1, p1Bytes[point2:]...)

	crossedByt2 := append([]byte{}, p2Bytes[:point1]...)
	crossedByt2 = append(crossedByt2, p1Bytes[point1:point2]...)
	crossedByt2 = append(crossedByt2, p2Bytes[point2:]...)

	return []byte(crossedByt1), []byte(crossedByt2)
}

func factorizationCrossover(a interface{}, b interface{}) (interface{}, interface{}) {
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
		var newV1 int
		var newV2 int
		divisor := int(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2
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
		var newV1 int8
		var newV2 int8
		divisor := int8(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2
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
		var newV1 int16
		var newV2 int16
		divisor := int16(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}

		return newV1, newV2
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
		var newV1 int32
		var newV2 int32
		divisor := int32(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

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
		var newV1 int64
		var newV2 int64
		divisor := RandRange(1, int64(min)/2)
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

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
		var newV1 uint
		var newV2 uint
		divisor := uint(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2
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
		var newV1 uint8
		var newV2 uint8
		divisor := uint8(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

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
		var newV1 uint16
		var newV2 uint16
		divisor := uint16(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

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
		var newV1 uint32
		var newV2 uint32
		divisor := uint32(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

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
		var newV1 uint64
		var newV2 uint64
		divisor := uint64(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case float32:
		v2, ok := b.(float32)
		if !ok {
			panic("type not equal")
		}
		newV1 := float32(0.7*float64(v1) + 0.3*float64(v2))
		newV2 := float32(0.3*float64(v1) + 0.7*float64(v2))
		return newV1, newV2
	case float64:
		v2, ok := b.(float64)
		if !ok {
			panic("type not equal")
		}
		newV1 := float64(0.3*float64(v1) + 0.7*float64(v2))
		newV2 := float64(0.3*float64(v1) + 0.7*float64(v2))
		return newV1, newV2
	default:
		panic("type can't be  factorization crossovered.")
	}
}

func randomByteFrom(seedByte byte) byte {
	p := GetSingleRand().Float64() // 0.0 <= p < 1.0

	var b uint8

	currentCase := determineCase(rune(seedByte))

	isProgress := GenerateRandomBool(0.5)
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
			isProgress = false
			break
		case p >= 0.1 && p < 0.25:
			currentCase = (currentCase + 3) % 4
			isProgress = false
			break
		case p >= 0.25 && p < 0.75:
			break

		case p >= 0.75 && p < 0.90:
			isProgress = true
			currentCase = (currentCase + 1) % 4
		case p >= 0.90 && p <= 1.0:
			isProgress = true
			currentCase = (currentCase + 2) % 4
		}
	}

	if currentCase == 2 {
		p2 := GetSingleRand().Float64()
		if isProgress {
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
		b = uint8(RandRange(AbsoluteMin, SpecialMAX+1))
	case 1:
		b = uint8(RandRange(LowAsciiMIN, LowAsciiMAX+1))
	case 2:

		b = Del
	case 3:
		b = uint8(RandRange(HighAsciiMin, HighAscillMax+1))

	}

	return byte(b)
}

const (
	AbsoluteMin = 0x00
	SpecialMAX  = 0x1F

	LowAsciiMIN = 0x20
	LowAsciiMAX = 0x7E
	Del         = 0x7F

	HighAsciiMin  = 0x80
	HighAscillMax = 0xFF

	PrintUnicodeMin = 0x100
	PrintUnicodeMax = 0xD7FF

	BoundaryUnicodeMIN = 0xD800
	BoundaryUnicodeMax = 0x10FFFF

	AbsoluteMax = 0x7FFFFFFF
)

func determineCase(seedRune rune) int {
	switch {
	case seedRune >= AbsoluteMin && seedRune <= SpecialMAX:
		return 0
	case seedRune >= LowAsciiMIN && seedRune <= LowAsciiMAX:
		return 1
	case seedRune == Del:
		return 2
	case seedRune >= HighAsciiMin && seedRune <= HighAscillMax:
		return 3
	case seedRune >= PrintUnicodeMin && seedRune <= PrintUnicodeMax:
		return 4
	case seedRune >= BoundaryUnicodeMIN && seedRune <= BoundaryUnicodeMax:
		return 5
	default:
		return 6
	}
}

func randomIntFrom(i interface{}) interface{} {
	p := GetSingleRand().Float64()

	switch v := i.(type) {
	case int:
		var i interface{}
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
			i = int(RandRange(min, max))
		case p < 0.3:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		case p < 0.45:
			min := int64(v) * (-8)
			max := int64(v) * (8)
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		case p < 0.60:
			min := int64(v) * (-16)
			max := int64(v) * (16)
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		default:
			i = GetSingleRand().Int()
		}
		return i

	case int8:
		if v == 0 {
			return int8(RandInt64())
		}
		var i8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			i8 = int8(RandRange(min, max))
		case p < 0.5:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i8 = int8(RandRange(min, max))
		default:
			i8 = int8(RandRange(-128, 128))
		}
		return i8

	case int16:
		if v == 0 {
			return int16(RandInt64())
		}
		var i16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			i16 = int16(RandRange(min, max))
		case p < 0.5:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i16 = int16(RandRange(min, max))
		default:
			i16 = int16(RandRange(-32768, 32768))
		}
		return i16

	case int32:
		if v == 0 {
			return int32(RandInt64())
		}
		var i16 interface{}
		switch {
		case p < 0.2:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			i16 = int32(RandRange(min, max))
		case p < 0.4:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i16 = int32(RandRange(min, max))
		case p < 0.6:
			min := int64(v) * (-8)
			max := int64(v) * (8)
			if min > max {
				min, max = max, min
			}
			i16 = int32(RandRange(min, max))
		default:
			i16 = GetSingleRand().Int32()
		}
		return i16

	case int64:
		if v == 0 {
			return RandInt64()
		}
		var i64 interface{}
		switch {
		case p < 0.15:
			min := v * (-2)
			max := v * (2)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		case p < 0.3:
			min := v * (-4)
			max := v * (4)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		case p < 0.45:
			min := v * (-8)
			max := v * (8)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		case p < 0.60:
			min := v * (-16)
			max := v * (16)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		default:
			i64 = GetSingleRand().Int64()
		}
		return i64

	default:
		panic("it's not supported int type")
	}
}

func randomUintFrom(u interface{}) interface{} {
	p := GetSingleRand().Float64()

	switch v := u.(type) {
	case uint:
		if v == 0 {
			return uint(RandUint64())
		}
		var u interface{}
		switch {
		case p < 0.3:
			min := v / 256
			max := v * 2
			u = uint(UintRandRange(uint64(min), uint64(max)))
		case p < 0.6:
			min := v / 9096
			max := v * 8
			u = uint(UintRandRange(uint64(min), uint64(max)))
		default:
			u = uint(GetSingleRand().Uint64())
		}
		return u

	case uint8:
		if v == 0 {
			return uint8(RandUint64())
		}
		var u8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 8
			max := int64(v) * 2
			u8 = uint8(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 32
			max := int64(v) * 8
			u8 = uint8(RandRange(min, max))
		default:
			u8 = uint8(RandRange(0, 256))
		}
		return u8

	case uint16:
		if v == 0 {
			return uint16(RandUint64())
		}
		var u16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 256
			max := int64(v) * 2
			u16 = uint16(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 9096
			max := int64(v) * 8
			u16 = uint16(RandRange(min, max))
		default:
			u16 = uint16(RandRange(0, 65536))
		}
		return u16

	case uint32:
		if v == 0 {
			return uint32(RandUint64())
		}
		var u32 interface{}
		switch {
		case p < 0.2:
			min := int64(v) / 256
			max := int64(v) * 2
			u32 = uint32(RandRange(min, max))
		case p < 0.4:
			min := int64(v) / 9096
			max := int64(v) * 8
			u32 = uint32(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / (9096 * 9096)
			max := int64(v) * 16
			u32 = uint32(RandRange(min, max))
		default:
			u32 = uint32(RandRange(0, int64(^uint32(0))))
		}
		return u32

	case uint64:
		if v == 0 {
			return RandUint64()
		}
		var u64 interface{}
		switch {
		case p < 0.2:
			min := v / 256
			max := v * 2
			u64 = UintRandRange(min, max)
		case p < 0.5:
			min := v / 9096
			max := v * 8
			u64 = UintRandRange(min, max)
		case p < 0.5:
			min := v / (9096 * 9096)
			max := v * 16
			u64 = UintRandRange(min, max)
		default:
			u64 = RandUint64()
		}
		return u64

	default:
		panic("it's not a supported uint type")
	}
}

func randomFloatFrom(f interface{}) interface{} {
	switch v := f.(type) {
	case float32:
		var f32 float32
		f32 = randFloat32From(float32(v))
		return float32(f32)
	case float64:
		var f64 float64
		f64 = randFloat64From(float64(v))
		return float64(f64)
	default:
		panic("argument is not float 32 or float4")
	}
}

func randFloat32From(f float32) float32 {
	return randomFloat32(f)
}

func randFloat64From(f float64) float64 {
	return randomFloat64(f)
}

func randomBool() bool {
	return UniformRandomBool(0.5)
}

func randomFloat32(a float32) float32 {
	bits := math.Float32bits(a)

	exponent := (bits >> 23) & 0xFF
	mantissa := bits & 0x7FFFFF
	sign := bits & 0x80000000
	t := uint32(unixNano())
	manshift := 1 + (t % 7)

	var shift int8
	if exponent <= 1 {
		shift = int8(1 + int(mantissa%2))
	} else if exponent >= 0xFE {
		shift = int8(-1 - int(mantissa%2))
	} else {
		shift = int8(-2 + int(mantissa%5))
	}

	newExp := int32(exponent) + int32(shift)
	newExponent := uint32(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 23) | (newMantissa & 0x7FFFFF)

	return math.Float32frombits(newBits)
}

func randomFloat64(a float64) float64 {
	bits := math.Float64bits(a)

	exponent := (bits >> 52) & 0x7FF

	mantissa := bits & 0xFFFFFFFFFFFFF

	sign := bits & 0x8000000000000000

	t := uint64(time.Now().UnixNano())
	manshift := 1 + (t % 7)

	var shift int16
	if exponent <= 1 {
		shift = int16(1 + int64(mantissa%2))
	} else if exponent >= 0x7FE {
		shift = int16(-1 - int64(mantissa%2))
	} else {
		shift = int16(-2 + int64(mantissa%5))
	}

	newExp := int64(exponent) + int64(shift)
	newExponent := uint64(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 52) | (newMantissa & 0xFFFFFFFFFFFFF)

	return math.Float64frombits(newBits)
}
