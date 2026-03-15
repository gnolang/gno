// Meta-benchmarks for the benchops measurement infrastructure itself.
// These verify that SwitchOpCode/finalizeCurrent overhead is negligible
// relative to the opcode benchmarks in bench_ops_test.go.
// Not related to gas calibration.
package benchops

import "testing"

func BenchmarkSwitchOpCode(b *testing.B) {
	InitMeasure()
	BeginOpCode(0x01)
	b.ResetTimer()
	for range b.N {
		SwitchOpCode(0x02)
	}
	b.StopTimer()
	StopOpCode()
}

func BenchmarkFinalizeCurrent(b *testing.B) {
	InitMeasure()
	measure.curOpCode = 0x01
	b.ResetTimer()
	for range b.N {
		measure.curStart = measure.timeZero // prevent accumulation
		finalizeCurrent()
	}
}
