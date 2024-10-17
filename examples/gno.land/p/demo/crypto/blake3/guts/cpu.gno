//go:build !darwin
// +build !darwin

package guts

// XXX: removed unused cpuid detection
// import "github.com/klauspost/cpuid/v2"

// var (
// 	haveAVX2   = cpuid.CPU.Supports(cpuid.AVX2)
// 	haveAVX512 = cpuid.CPU.Supports(cpuid.AVX512F)
// )

// XXX: we don't need to initialize these variables since we will always use the default values (false)
const (
	haveAVX2   bool = false
	haveAVX512 bool = false
)
