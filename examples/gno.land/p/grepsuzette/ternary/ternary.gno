package ternary

import "std"

func String(cond bool, yes, no string) string {
	if cond {
		return yes
	}
	return no
}

func Int(cond bool, yes, no int) int {
	if cond {
		return yes
	}
	return no
}

func Int8(cond bool, yes, no int8) int8 {
	if cond {
		return yes
	}
	return no
}

func Int16(cond bool, yes, no int16) int16 {
	if cond {
		return yes
	}
	return no
}

func Int32(cond bool, yes, no int32) int32 {
	if cond {
		return yes
	}
	return no
}

func Int64(cond bool, yes, no int64) int64 {
	if cond {
		return yes
	}
	return no
}

func Uint(cond bool, yes, no uint) uint {
	if cond {
		return yes
	}
	return no
}

func Uint8(cond bool, yes, no uint8) uint8 {
	if cond {
		return yes
	}
	return no
}

func Uint16(cond bool, yes, no uint16) uint16 {
	if cond {
		return yes
	}
	return no
}

func Uint32(cond bool, yes, no uint32) uint32 {
	if cond {
		return yes
	}
	return no
}

func Uint64(cond bool, yes, no uint64) uint64 {
	if cond {
		return yes
	}
	return no
}

func Float32(cond bool, yes, no float32) float32 {
	if cond {
		return yes
	}
	return no
}

func Float64(cond bool, yes, no float64) float64 {
	if cond {
		return yes
	}
	return no
}

func Rune(cond bool, yes, no rune) rune {
	if cond {
		return yes
	}
	return no
}

func Bool(cond bool, yes, no bool) bool {
	if cond {
		return yes
	}
	return no
}

func Address(cond bool, yes, no std.Address) std.Address {
	if cond {
		return yes
	}
	return no
}
