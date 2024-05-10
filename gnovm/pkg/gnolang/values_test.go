package gnolang

import (
	"math"
	"testing"
)

func BenchmarkTVGetBool(b *testing.B) {
	var tv TypedValue
	tv.N[0] = 1
	for i := 0; i < b.N; i++ {
		tv.GetBool()
	}
}

func BenchmarkTVGetBoolNew(b *testing.B) {
	var tv TypedValue
	tv.N[0] = 1
	for i := 0; i < b.N; i++ {
		tv.GetBoolNew()
	}
}

func BenchmarkTVSetBool(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetBool(true)
	}
}

func BenchmarkTVSetBoolNew(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetBoolNew(true)
	}
}

func BenchmarkTVGetInt(b *testing.B) {
	var tv TypedValue
	tv.SetInt(math.MaxInt64)
	for i := 0; i < b.N; i++ {
		tv.GetInt()
	}
}

func BenchmarkTVGetIntNew(b *testing.B) {
	var tv TypedValue
	tv.SetIntNew(math.MaxInt64)
	for i := 0; i < b.N; i++ {
		tv.GetIntNew()
	}
}

func BenchmarkTVSetInt(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetInt(math.MaxInt64)
	}
}

func BenchmarkTVSetIntNew(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetIntNew(math.MaxInt64)
	}
}

func BenchmarkTVSetInt16(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetInt16(math.MaxInt16)
	}
}

func BenchmarkTVSetInt16New(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetInt16New(math.MaxInt16)
	}
}

func BenchmarkTVGetInt16(b *testing.B) {
	var tv TypedValue
	tv.SetInt16(math.MaxInt16)
	for i := 0; i < b.N; i++ {
		tv.GetInt16()
	}
}

func BenchmarkTVGetInt16New(b *testing.B) {
	var tv TypedValue
	tv.SetInt16New(math.MaxInt16)
	for i := 0; i < b.N; i++ {
		tv.GetInt16New()
	}
}

func BenchmarkTVSetFloat64(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetFloat64(math.MaxFloat64)
	}
}

func BenchmarkTVSetFloat64New(b *testing.B) {
	var tv TypedValue
	for i := 0; i < b.N; i++ {
		tv.SetFloat64New(math.MaxFloat64)
	}
}

func BenchmarkTVGetFloat64(b *testing.B) {
	var tv TypedValue
	tv.SetFloat64(math.MaxFloat64)
	for i := 0; i < b.N; i++ {
		tv.GetFloat64()
	}
}

func BenchmarkTVGetFloat64New(b *testing.B) {
	var tv TypedValue
	tv.SetFloat64New(math.MaxFloat64)
	for i := 0; i < b.N; i++ {
		tv.GetFloat64New()
	}
}
