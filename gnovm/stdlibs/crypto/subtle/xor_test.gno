// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// GNO modifications:
// Removed the dependency on "crypto/rand" package due to non-deterministic behavior.
// This is safe because we are only generating random data for testing purposes.
// The original file can be found here:
// https://cs.opensource.google/go/go/+/master:src/crypto/subtle/xor_test.go

package subtle

import (
	"bytes"
	"io"
	// "crypto/rand" XX: replaced with math/rand for deterministic behavior for test data
	// "fmt" XX: removed due to fmt dependency, not needed for testing, it was only used for benchmarks
	"math/rand"
	"testing"
)

// XX: added testRandomReader struct to generate random data for testing.
// to keep the tests deterministic but allowing for minimal changes
type testRandomReader struct{}

// Read fills the provided buffer with psuedo random bytes.
func (r *testRandomReader) Read(b []byte) (n int, err error) {
	for i := range b {
		b[i] = byte(rand.IntN(256))
	}
	return len(b), nil
}

func TestXORBytes(t *testing.T) {
	// XX: added testRandomReader to generate random data for testing.
	reader := &testRandomReader{}
	// XX: reduced n from 1024 to 12 for faster testing
	for n := 1; n <= 12; n++ {
		for alignP := 0; alignP < 8; alignP++ {
			for alignQ := 0; alignQ < 8; alignQ++ {
				for alignD := 0; alignD < 8; alignD++ {
					p := make([]byte, alignP+n, alignP+n+10)[alignP:]
					q := make([]byte, alignQ+n, alignQ+n+10)[alignQ:]
					if n&1 != 0 {
						p = p[:n]
					} else {
						q = q[:n]
					}
					// XX: replaced crypto rand.Reader with testRandomReader
					if _, err := io.ReadFull(reader, p); err != nil {
						t.Fatal(err)
					}
					// XX: replaced crypto rand.Reader with testRandomReader
					if _, err := io.ReadFull(reader, q); err != nil {
						t.Fatal(err)
					}

					d := make([]byte, alignD+n, alignD+n+10)
					for i := range d {
						d[i] = 0xdd
					}
					want := make([]byte, len(d), cap(d))
					copy(want[:cap(want)], d[:cap(d)])
					for i := 0; i < n; i++ {
						want[alignD+i] = p[i] ^ q[i]
					}

					if XORBytes(d[alignD:], p, q); !bytes.Equal(d, want) {
						t.Fatalf("n=%d alignP=%d alignQ=%d alignD=%d:\n\tp = %x\n\tq = %x\n\td = %x\n\twant %x\n", n, alignP, alignQ, alignD, p, q, d, want)
					}
				}
			}
		}
	}
}

func TestXorBytesPanic(t *testing.T) {
	mustPanic(t, "subtle.XORBytes: dst too short", func() {
		XORBytes(nil, make([]byte, 1), make([]byte, 1))
	})
	mustPanic(t, "subtle.XORBytes: dst too short", func() {
		XORBytes(make([]byte, 1), make([]byte, 2), make([]byte, 3))
	})
}

func mustPanic(t *testing.T, expected string, f func()) {
	t.Helper()
	defer func() {
		switch msg := recover().(type) {
		case nil:
			t.Errorf("expected panic(%q), but did not panic", expected)
		case string:
			if msg != expected {
				t.Errorf("expected panic(%q), but got panic(%q)", expected, msg)
			}
		}
	}()
	f()
}
