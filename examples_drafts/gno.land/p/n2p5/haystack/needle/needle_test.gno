package needle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestNeedle(t *testing.T) {
	t.Parallel()
	t.Run("Bytes", func(t *testing.T) {
		t.Parallel()
		p, _ := hex.DecodeString("40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")
		n, _ := New(p)
		b := n.Bytes()
		b[0], b[1], b[2], b[3] = 0, 0, 0, 0
		if bytes.Equal(n.Bytes(), b) {
			t.Error("mutating Bytes() changed needle bytes")
		}
	})
	t.Run("Payload", func(t *testing.T) {
		t.Parallel()
		p, _ := hex.DecodeString("40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")
		n, _ := New(p)
		payload := n.Payload()
		if !bytes.Equal(p, payload) {
			t.Error("payload imported by New does not match needle.Payload()")
		}
		payload[0] = 0
		pl := n.Payload()
		if bytes.Equal(pl, payload) {
			t.Error("mutating Payload() changed needle payload")
		}
	})
	t.Run("Hash", func(t *testing.T) {
		t.Parallel()
		p, _ := hex.DecodeString("40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")
		n, _ := New(p)
		hash := n.Hash()
		h := sha256.Sum256(p)
		if !bytes.Equal(h[:], hash) {
			t.Error("exported hash is invalid")
		}
		hash[0] = 0
		h2 := n.Hash()
		if bytes.Equal(h2, hash) {
			t.Error("mutating Hash() changed needle hash")
		}
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	p, _ := hex.DecodeString("40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")
	expected, _ := hex.DecodeString("f1b462c84a0c51dad44293951f0b084a8871b3700ac1b9fc7a53a20bc0ba0fed40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")

	testTable := []struct {
		payload     []byte
		expected    []byte
		hasError    bool
		description string
	}{
		{
			payload:     p,
			expected:    expected,
			hasError:    false,
			description: "expected payload",
		},
		{
			payload:     p[:PayloadLength-1],
			expected:    nil,
			hasError:    true,
			description: "payload invalid length (too small)",
		},
		{
			payload:     append(p, byte(1)),
			expected:    nil,
			hasError:    true,
			description: "payload invalid length (too large)",
		},
	}

	for _, test := range testTable {
		n, err := New(test.payload)
		if err != nil {
			if !test.hasError {
				t.Errorf("test: %v had error: %v", test.description, err)
			}
		} else if !bytes.Equal(n.Bytes(), test.expected) {
			t.Errorf("%v, bytes not equal\n%x\n%x", test.description, n.Bytes(), test.expected)
		}
	}
}

func TestFromBytes(t *testing.T) {
	t.Parallel()

	validRaw, _ := hex.DecodeString("f1b462c84a0c51dad44293951f0b084a8871b3700ac1b9fc7a53a20bc0ba0fed40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")
	validExpected, _ := hex.DecodeString("f1b462c84a0c51dad44293951f0b084a8871b3700ac1b9fc7a53a20bc0ba0fed40e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")
	invalidHash, _ := hex.DecodeString("182e0ca0d2fb1da76da6caf36a9d0d2838655632e85891216dc8b545d8f1410940e4350b03d8b0c9e340321210b259d9a20b19632929b4a219254a4269c11f820c75168c6a91d309f4b134a7d715a5ac408991e1cf9415995053cf8a4e185dae22a06617ac51ebf7d232bc49e567f90be4db815c2b88ca0d9a4ef7a5119c0e592c88dfb96706e6510fb8a657c0f70f6695ea310d24786e6d980e9b33cf2665342b965b2391f6bb982c4c5f6058b9cba58038d32452e07cdee9420a8bd7f514e1")

	testTable := []struct {
		rawBytes    []byte
		expected    []byte
		hasError    bool
		description string
	}{
		{
			rawBytes:    validRaw,
			expected:    validExpected,
			hasError:    false,
			description: "valid raw bytes",
		},
		{
			rawBytes:    make([]byte, 0),
			expected:    nil,
			hasError:    true,
			description: "empty bytes",
		},
		{
			rawBytes:    make([]byte, NeedleLength-1),
			expected:    nil,
			hasError:    true,
			description: "too few bytes, one less than expected",
		},
		{
			rawBytes:    make([]byte, 0),
			expected:    nil,
			hasError:    true,
			description: "too few bytes, no bytes",
		},
		{
			rawBytes:    make([]byte, NeedleLength+1),
			expected:    nil,
			hasError:    true,
			description: "too many bytes",
		},
		{
			rawBytes:    invalidHash,
			expected:    nil,
			hasError:    true,
			description: "invalid hash",
		},
	}
	for _, test := range testTable {
		n, err := FromBytes(test.rawBytes)
		if err != nil {
			if !test.hasError {
				t.Errorf("test: %v had error: %v", test.description, err)
			}
		} else if !bytes.Equal(n.Bytes(), test.expected) {
			t.Errorf("%v, bytes not equal\n%x\n%x", test.description, n.Bytes(), test.expected)
		}
	}
}
