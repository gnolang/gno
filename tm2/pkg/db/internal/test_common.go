package internal

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
)

// RandStr constructs a random alphanumeric string of given length.
func RandStr(length int) string {
	chars := []byte{}
MAIN_LOOP:
	for {
		//nolint:gosec
		val := rand.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}

	return string(chars)
}

// ----------------------------------------
// MockIterator

type MockIterator struct{}

func (MockIterator) Domain() (start []byte, end []byte) {
	return nil, nil
}

func (MockIterator) Valid() bool {
	return false
}

func (MockIterator) Next() {
}

func (MockIterator) Key() []byte {
	return nil
}

func (MockIterator) Value() []byte {
	return nil
}

func (MockIterator) Close() {
}

func BenchmarkRandomReadsWrites(b *testing.B, db db.DB) {
	b.Helper()

	b.StopTimer()

	// create dummy data
	const numItems = int64(1000000)
	internal := map[int64]int64{}
	for i := 0; i < int(numItems); i++ {
		internal[int64(i)] = int64(0)
	}

	// fmt.Println("ok, starting")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// Write something
		{
			//nolint:gosec
			idx := int64(rand.Int()) % numItems
			internal[idx]++
			val := internal[idx]
			idxBytes := int642Bytes(idx)
			valBytes := int642Bytes(val)
			// fmt.Printf("Set %X -> %X\n", idxBytes, valBytes)
			db.Set(idxBytes, valBytes)
		}

		// Read something
		{
			//nolint:gosec
			idx := int64(rand.Int()) % numItems
			valExp := internal[idx]
			idxBytes := int642Bytes(idx)
			valBytes := db.Get(idxBytes)
			// fmt.Printf("Get %X -> %X\n", idxBytes, valBytes)
			if valExp == 0 {
				if !bytes.Equal(valBytes, nil) {
					b.Errorf("Expected %v for %v, got %X", nil, idx, valBytes)
					break
				}
			} else {
				if len(valBytes) != 8 {
					b.Errorf("Expected length 8 for %v, got %X", idx, valBytes)
					break
				}
				valGot := bytes2Int64(valBytes)
				if valExp != valGot {
					b.Errorf("Expected %v for %v, got %v", valExp, idx, valGot)
					break
				}
			}
		}
	}
}

func int642Bytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytes2Int64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
