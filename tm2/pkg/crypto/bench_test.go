package crypto

import (
	"crypto/rand"
	"testing"
)

var sink any = nil

func BenchmarkAddressCompare(b *testing.B) {
	var addr1, addr2 Address
	rand.Read(addr1[:])
	rand.Read(addr2[:])
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sink = addr1.Compare(addr1)
		sink = addr1.Compare(addr2)
		sink = addr2.Compare(addr2)
		sink = addr2.Compare(addr1)
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}

	sink = nil
}
