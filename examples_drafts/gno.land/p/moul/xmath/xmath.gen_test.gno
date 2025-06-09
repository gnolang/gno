package xmath

import "testing"

func TestInt8Helpers(t *testing.T) {
	// Test MaxInt8
	if MaxInt8(1, 2) != 2 {
		t.Error("MaxInt8(1, 2) should be 2")
	}
	if MaxInt8(-1, -2) != -1 {
		t.Error("MaxInt8(-1, -2) should be -1")
	}

	// Test MinInt8
	if MinInt8(1, 2) != 1 {
		t.Error("MinInt8(1, 2) should be 1")
	}
	if MinInt8(-1, -2) != -2 {
		t.Error("MinInt8(-1, -2) should be -2")
	}

	// Test ClampInt8
	if ClampInt8(5, 1, 3) != 3 {
		t.Error("ClampInt8(5, 1, 3) should be 3")
	}
	if ClampInt8(0, 1, 3) != 1 {
		t.Error("ClampInt8(0, 1, 3) should be 1")
	}
	if ClampInt8(2, 1, 3) != 2 {
		t.Error("ClampInt8(2, 1, 3) should be 2")
	}

	// Test AbsInt8
	if AbsInt8(-5) != 5 {
		t.Error("AbsInt8(-5) should be 5")
	}
	if AbsInt8(5) != 5 {
		t.Error("AbsInt8(5) should be 5")
	}

	// Test SignInt8
	if SignInt8(-5) != -1 {
		t.Error("SignInt8(-5) should be -1")
	}
	if SignInt8(5) != 1 {
		t.Error("SignInt8(5) should be 1")
	}
	if SignInt8(0) != 0 {
		t.Error("SignInt8(0) should be 0")
	}

}

func TestInt16Helpers(t *testing.T) {
	// Test MaxInt16
	if MaxInt16(1, 2) != 2 {
		t.Error("MaxInt16(1, 2) should be 2")
	}
	if MaxInt16(-1, -2) != -1 {
		t.Error("MaxInt16(-1, -2) should be -1")
	}

	// Test MinInt16
	if MinInt16(1, 2) != 1 {
		t.Error("MinInt16(1, 2) should be 1")
	}
	if MinInt16(-1, -2) != -2 {
		t.Error("MinInt16(-1, -2) should be -2")
	}

	// Test ClampInt16
	if ClampInt16(5, 1, 3) != 3 {
		t.Error("ClampInt16(5, 1, 3) should be 3")
	}
	if ClampInt16(0, 1, 3) != 1 {
		t.Error("ClampInt16(0, 1, 3) should be 1")
	}
	if ClampInt16(2, 1, 3) != 2 {
		t.Error("ClampInt16(2, 1, 3) should be 2")
	}

	// Test AbsInt16
	if AbsInt16(-5) != 5 {
		t.Error("AbsInt16(-5) should be 5")
	}
	if AbsInt16(5) != 5 {
		t.Error("AbsInt16(5) should be 5")
	}

	// Test SignInt16
	if SignInt16(-5) != -1 {
		t.Error("SignInt16(-5) should be -1")
	}
	if SignInt16(5) != 1 {
		t.Error("SignInt16(5) should be 1")
	}
	if SignInt16(0) != 0 {
		t.Error("SignInt16(0) should be 0")
	}

}

func TestInt32Helpers(t *testing.T) {
	// Test MaxInt32
	if MaxInt32(1, 2) != 2 {
		t.Error("MaxInt32(1, 2) should be 2")
	}
	if MaxInt32(-1, -2) != -1 {
		t.Error("MaxInt32(-1, -2) should be -1")
	}

	// Test MinInt32
	if MinInt32(1, 2) != 1 {
		t.Error("MinInt32(1, 2) should be 1")
	}
	if MinInt32(-1, -2) != -2 {
		t.Error("MinInt32(-1, -2) should be -2")
	}

	// Test ClampInt32
	if ClampInt32(5, 1, 3) != 3 {
		t.Error("ClampInt32(5, 1, 3) should be 3")
	}
	if ClampInt32(0, 1, 3) != 1 {
		t.Error("ClampInt32(0, 1, 3) should be 1")
	}
	if ClampInt32(2, 1, 3) != 2 {
		t.Error("ClampInt32(2, 1, 3) should be 2")
	}

	// Test AbsInt32
	if AbsInt32(-5) != 5 {
		t.Error("AbsInt32(-5) should be 5")
	}
	if AbsInt32(5) != 5 {
		t.Error("AbsInt32(5) should be 5")
	}

	// Test SignInt32
	if SignInt32(-5) != -1 {
		t.Error("SignInt32(-5) should be -1")
	}
	if SignInt32(5) != 1 {
		t.Error("SignInt32(5) should be 1")
	}
	if SignInt32(0) != 0 {
		t.Error("SignInt32(0) should be 0")
	}

}

func TestInt64Helpers(t *testing.T) {
	// Test MaxInt64
	if MaxInt64(1, 2) != 2 {
		t.Error("MaxInt64(1, 2) should be 2")
	}
	if MaxInt64(-1, -2) != -1 {
		t.Error("MaxInt64(-1, -2) should be -1")
	}

	// Test MinInt64
	if MinInt64(1, 2) != 1 {
		t.Error("MinInt64(1, 2) should be 1")
	}
	if MinInt64(-1, -2) != -2 {
		t.Error("MinInt64(-1, -2) should be -2")
	}

	// Test ClampInt64
	if ClampInt64(5, 1, 3) != 3 {
		t.Error("ClampInt64(5, 1, 3) should be 3")
	}
	if ClampInt64(0, 1, 3) != 1 {
		t.Error("ClampInt64(0, 1, 3) should be 1")
	}
	if ClampInt64(2, 1, 3) != 2 {
		t.Error("ClampInt64(2, 1, 3) should be 2")
	}

	// Test AbsInt64
	if AbsInt64(-5) != 5 {
		t.Error("AbsInt64(-5) should be 5")
	}
	if AbsInt64(5) != 5 {
		t.Error("AbsInt64(5) should be 5")
	}

	// Test SignInt64
	if SignInt64(-5) != -1 {
		t.Error("SignInt64(-5) should be -1")
	}
	if SignInt64(5) != 1 {
		t.Error("SignInt64(5) should be 1")
	}
	if SignInt64(0) != 0 {
		t.Error("SignInt64(0) should be 0")
	}

}

func TestIntHelpers(t *testing.T) {
	// Test MaxInt
	if MaxInt(1, 2) != 2 {
		t.Error("MaxInt(1, 2) should be 2")
	}
	if MaxInt(-1, -2) != -1 {
		t.Error("MaxInt(-1, -2) should be -1")
	}

	// Test MinInt
	if MinInt(1, 2) != 1 {
		t.Error("MinInt(1, 2) should be 1")
	}
	if MinInt(-1, -2) != -2 {
		t.Error("MinInt(-1, -2) should be -2")
	}

	// Test ClampInt
	if ClampInt(5, 1, 3) != 3 {
		t.Error("ClampInt(5, 1, 3) should be 3")
	}
	if ClampInt(0, 1, 3) != 1 {
		t.Error("ClampInt(0, 1, 3) should be 1")
	}
	if ClampInt(2, 1, 3) != 2 {
		t.Error("ClampInt(2, 1, 3) should be 2")
	}

	// Test AbsInt
	if AbsInt(-5) != 5 {
		t.Error("AbsInt(-5) should be 5")
	}
	if AbsInt(5) != 5 {
		t.Error("AbsInt(5) should be 5")
	}

	// Test SignInt
	if SignInt(-5) != -1 {
		t.Error("SignInt(-5) should be -1")
	}
	if SignInt(5) != 1 {
		t.Error("SignInt(5) should be 1")
	}
	if SignInt(0) != 0 {
		t.Error("SignInt(0) should be 0")
	}

}

func TestUint8Helpers(t *testing.T) {
	// Test MaxUint8
	if MaxUint8(1, 2) != 2 {
		t.Error("MaxUint8(1, 2) should be 2")
	}

	// Test MinUint8
	if MinUint8(1, 2) != 1 {
		t.Error("MinUint8(1, 2) should be 1")
	}

	// Test ClampUint8
	if ClampUint8(5, 1, 3) != 3 {
		t.Error("ClampUint8(5, 1, 3) should be 3")
	}
	if ClampUint8(0, 1, 3) != 1 {
		t.Error("ClampUint8(0, 1, 3) should be 1")
	}
	if ClampUint8(2, 1, 3) != 2 {
		t.Error("ClampUint8(2, 1, 3) should be 2")
	}

}

func TestUint16Helpers(t *testing.T) {
	// Test MaxUint16
	if MaxUint16(1, 2) != 2 {
		t.Error("MaxUint16(1, 2) should be 2")
	}

	// Test MinUint16
	if MinUint16(1, 2) != 1 {
		t.Error("MinUint16(1, 2) should be 1")
	}

	// Test ClampUint16
	if ClampUint16(5, 1, 3) != 3 {
		t.Error("ClampUint16(5, 1, 3) should be 3")
	}
	if ClampUint16(0, 1, 3) != 1 {
		t.Error("ClampUint16(0, 1, 3) should be 1")
	}
	if ClampUint16(2, 1, 3) != 2 {
		t.Error("ClampUint16(2, 1, 3) should be 2")
	}

}

func TestUint32Helpers(t *testing.T) {
	// Test MaxUint32
	if MaxUint32(1, 2) != 2 {
		t.Error("MaxUint32(1, 2) should be 2")
	}

	// Test MinUint32
	if MinUint32(1, 2) != 1 {
		t.Error("MinUint32(1, 2) should be 1")
	}

	// Test ClampUint32
	if ClampUint32(5, 1, 3) != 3 {
		t.Error("ClampUint32(5, 1, 3) should be 3")
	}
	if ClampUint32(0, 1, 3) != 1 {
		t.Error("ClampUint32(0, 1, 3) should be 1")
	}
	if ClampUint32(2, 1, 3) != 2 {
		t.Error("ClampUint32(2, 1, 3) should be 2")
	}

}

func TestUint64Helpers(t *testing.T) {
	// Test MaxUint64
	if MaxUint64(1, 2) != 2 {
		t.Error("MaxUint64(1, 2) should be 2")
	}

	// Test MinUint64
	if MinUint64(1, 2) != 1 {
		t.Error("MinUint64(1, 2) should be 1")
	}

	// Test ClampUint64
	if ClampUint64(5, 1, 3) != 3 {
		t.Error("ClampUint64(5, 1, 3) should be 3")
	}
	if ClampUint64(0, 1, 3) != 1 {
		t.Error("ClampUint64(0, 1, 3) should be 1")
	}
	if ClampUint64(2, 1, 3) != 2 {
		t.Error("ClampUint64(2, 1, 3) should be 2")
	}

}

func TestUintHelpers(t *testing.T) {
	// Test MaxUint
	if MaxUint(1, 2) != 2 {
		t.Error("MaxUint(1, 2) should be 2")
	}

	// Test MinUint
	if MinUint(1, 2) != 1 {
		t.Error("MinUint(1, 2) should be 1")
	}

	// Test ClampUint
	if ClampUint(5, 1, 3) != 3 {
		t.Error("ClampUint(5, 1, 3) should be 3")
	}
	if ClampUint(0, 1, 3) != 1 {
		t.Error("ClampUint(0, 1, 3) should be 1")
	}
	if ClampUint(2, 1, 3) != 2 {
		t.Error("ClampUint(2, 1, 3) should be 2")
	}

}

func TestFloat32Helpers(t *testing.T) {
	// Test MaxFloat32
	if MaxFloat32(1, 2) != 2 {
		t.Error("MaxFloat32(1, 2) should be 2")
	}
	if MaxFloat32(-1, -2) != -1 {
		t.Error("MaxFloat32(-1, -2) should be -1")
	}

	// Test MinFloat32
	if MinFloat32(1, 2) != 1 {
		t.Error("MinFloat32(1, 2) should be 1")
	}
	if MinFloat32(-1, -2) != -2 {
		t.Error("MinFloat32(-1, -2) should be -2")
	}

	// Test ClampFloat32
	if ClampFloat32(5, 1, 3) != 3 {
		t.Error("ClampFloat32(5, 1, 3) should be 3")
	}
	if ClampFloat32(0, 1, 3) != 1 {
		t.Error("ClampFloat32(0, 1, 3) should be 1")
	}
	if ClampFloat32(2, 1, 3) != 2 {
		t.Error("ClampFloat32(2, 1, 3) should be 2")
	}

	// Test AbsFloat32
	if AbsFloat32(-5) != 5 {
		t.Error("AbsFloat32(-5) should be 5")
	}
	if AbsFloat32(5) != 5 {
		t.Error("AbsFloat32(5) should be 5")
	}

	// Test SignFloat32
	if SignFloat32(-5) != -1 {
		t.Error("SignFloat32(-5) should be -1")
	}
	if SignFloat32(5) != 1 {
		t.Error("SignFloat32(5) should be 1")
	}
	if SignFloat32(0.0) != 0 {
		t.Error("SignFloat32(0.0) should be 0")
	}

}

func TestFloat64Helpers(t *testing.T) {
	// Test MaxFloat64
	if MaxFloat64(1, 2) != 2 {
		t.Error("MaxFloat64(1, 2) should be 2")
	}
	if MaxFloat64(-1, -2) != -1 {
		t.Error("MaxFloat64(-1, -2) should be -1")
	}

	// Test MinFloat64
	if MinFloat64(1, 2) != 1 {
		t.Error("MinFloat64(1, 2) should be 1")
	}
	if MinFloat64(-1, -2) != -2 {
		t.Error("MinFloat64(-1, -2) should be -2")
	}

	// Test ClampFloat64
	if ClampFloat64(5, 1, 3) != 3 {
		t.Error("ClampFloat64(5, 1, 3) should be 3")
	}
	if ClampFloat64(0, 1, 3) != 1 {
		t.Error("ClampFloat64(0, 1, 3) should be 1")
	}
	if ClampFloat64(2, 1, 3) != 2 {
		t.Error("ClampFloat64(2, 1, 3) should be 2")
	}

	// Test AbsFloat64
	if AbsFloat64(-5) != 5 {
		t.Error("AbsFloat64(-5) should be 5")
	}
	if AbsFloat64(5) != 5 {
		t.Error("AbsFloat64(5) should be 5")
	}

	// Test SignFloat64
	if SignFloat64(-5) != -1 {
		t.Error("SignFloat64(-5) should be -1")
	}
	if SignFloat64(5) != 1 {
		t.Error("SignFloat64(5) should be 1")
	}
	if SignFloat64(0.0) != 0 {
		t.Error("SignFloat64(0.0) should be 0")
	}

}
