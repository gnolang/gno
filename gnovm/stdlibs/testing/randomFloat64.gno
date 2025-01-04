package testing

import (
	"math" // float32Bit 있음 파싱하는 거 참고해서 추가하셈
	"time"
)

func randomFloat32(a float32) float32 {
	// float32를 uint32로 변환하여 비트 조작
	bits := math.Float32bits(a)

	// 지수부 추출 (비트 23-30)
	exponent := (bits >> 23) & 0xFF
	// 가수부 추출 (비트 0-22)
	mantissa := bits & 0x7FFFFF
	// 부호 비트 추출
	sign := bits & 0x80000000

	// 시간값을 시드로 사용해 변동폭 생성
	// 시간 값 수정했음. time대신 해당 피키지의 함수 사용.
	t := uint32(unixNano())
	manshift := 1 + (t % 7) // 1~7 사이의 시프트 값 생성

	// 지수부 시프트 방향 결정
	// 시ㅂ프트 언더플로 막는 걸로 로직 바꿨음
	var shift int8
	if exponent <= 1 { // 지수가 매우 작을 때
		shift = int8(1 + int(mantissa%2)) // +1 또는 +2
	} else if exponent >= 0xFE { // 지수가 매우 클 때
		shift = int8(-1 - int(mantissa%2)) // -1 또는 -2
	} else { // 일반적인 경우
		shift = int8(-2 + int(mantissa%5)) // -2 ~ +2
	}

	newExp := int32(exponent) + int32(shift)
	newExponent := uint32(newExp)

	// 가수부 변경: 시프트 값을 랜덤하게 적용
	newMantissa := mantissa ^ (mantissa >> manshift)

	// 새로운 float32 조립
	newBits := sign | (newExponent << 23) | (newMantissa & 0x7FFFFF)

	return math.Float32frombits(newBits)
}

func randomFloat64(a float64) float64 {
	// float64를 uint64로 변환하여 비트 조작
	bits := math.Float64bits(a)

	// 지수부 추출 (비트 52-62)
	exponent := (bits >> 52) & 0x7FF
	// 가수부 추출 (비트 0-51)
	mantissa := bits & 0xFFFFFFFFFFFFF
	// 부호 비트 추출
	sign := bits & 0x8000000000000000

	// 시간값을 시드로 사용해 변동폭 생성
	t := uint64(time.Now().UnixNano())
	manshift := 1 + (t % 7) // 1~7 사이의 시프트 값 생성

	// 지수부 시프트 방향 결정
	// 형변환 추가함
	var shift int16
	if exponent <= 1 { // 지수가 매우 작을 때
		shift = int16(1 + int64(mantissa%2)) // +1 또는 +2
	} else if exponent >= 0x7FE { // 지수가 매우 클 때
		shift = int16(-1 - int64(mantissa%2)) // -1 또는 -2
	} else { // 일반적인 경우
		shift = int16(-2 + int64(mantissa%5)) // -2 ~ +2
	}

	newExp := int64(exponent) + int64(shift)
	newExponent := uint64(newExp)

	// 가수부 변경: 시프트 값을 랜덤하게 적용. 얘는 크게 언더오버같은 거 안남 xor 비트연산이여서.
	newMantissa := mantissa ^ (mantissa >> manshift)

	// 새로운 float64 조립
	newBits := sign | (newExponent << 52) | (newMantissa & 0xFFFFFFFFFFFFF)

	return math.Float64frombits(newBits)
}

//
// func main() {
//     initial := float32(123.456)
//     for i := 0; i < 5; i++ {
//         result := randomFloat32(initial)
//         fmt.Printf("Original: %f, Random: %f\n", initial, result)
//     }
// }
