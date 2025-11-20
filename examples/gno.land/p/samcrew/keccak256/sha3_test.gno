// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package keccak256

// Tests include all the ShortMsgKATs provided by the Keccak team at
// https://github.com/gvanas/KeccakCodePackage
//
// They only include the zero-bit case of the bitwise testvectors
// published by NIST in the draft of FIPS-202.

import (
	"bytes"
	"encoding/hex"
	"hash"
	"math/rand"
	"testing"
)

const (
	testString  = "brekeccakkeccak koax koax"
	katFilename = "testdata/keccakKats.json.deflate"
)

// testDigests contains functions returning hash.Hash instances
// with output-length equal to the KAT length for SHA-3, Keccak
// and SHAKE instances.
var testDigests = map[string]func() hash.Hash{
	"Keccak-256": NewLegacyKeccak256,
}

// decodeHex converts a hex-encoded string into a raw byte string.
func decodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// structs used to marshal JSON test-cases.
type KeccakKats struct {
	Kats map[string][]struct {
		Digest  string `json:"digest"`
		Length  int64  `json:"length"`
		Message string `json:"message"`

		// Defined only for cSHAKE
		N string `json:"N"`
		S string `json:"S"`
	}
}

// TestKeccakKats tests the SHA-3 and Shake implementations against all the
// ShortMsgKATs from https://github.com/gvanas/KeccakCodePackage
// (The testvectors are stored in keccakKats.json.deflate due to their length.)
//XXX Do not use "compress/flate"
/*func TestKeccakKats(t *testing.T) {
	// Read the KATs.
	deflated, err := os.Open(katFilename)
	if err != nil {
		t.Errorf("error opening %s: %s", katFilename, err)
	}
	file := flate.NewReader(deflated)
	dec := json.NewDecoder(file)
	var katSet KeccakKats
	err = dec.Decode(&katSet)
	if err != nil {
		t.Errorf("error decoding KATs: %s", err)
	}

	for algo, function := range testDigests {
		d := function()
		for _, kat := range katSet.Kats[algo] {
			d.Reset()
			in, err := hex.DecodeString(kat.Message)
			if err != nil {
				t.Errorf("error decoding KAT: %s", err)
			}
			d.Write(in[:kat.Length/8])
			got := strings.ToUpper(hex.EncodeToString(d.Sum(nil)))
			if got != kat.Digest {
				t.Errorf("function=%s, length=%d\nmessage:\n %s\ngot:\n  %s\nwanted:\n %s",
					algo, kat.Length, kat.Message, got, kat.Digest)
				t.Logf("wanted %+v", kat)
				t.FailNow()
			}
			continue
		}
	}

	for algo, v := range testShakes {
		for _, kat := range katSet.Kats[algo] {
			N, err := hex.DecodeString(kat.N)
			if err != nil {
				t.Errorf("error decoding KAT: %s", err)
			}

			S, err := hex.DecodeString(kat.S)
			if err != nil {
				t.Errorf("error decoding KAT: %s", err)
			}
			d := v.constructor(N, S)
			in, err := hex.DecodeString(kat.Message)
			if err != nil {
				t.Errorf("error decoding KAT: %s", err)
			}

			d.Write(in[:kat.Length/8])
			out := make([]byte, len(kat.Digest)/2)
			d.Read(out)
			got := strings.ToUpper(hex.EncodeToString(out))
			if got != kat.Digest {
				t.Errorf("function=%s, length=%d N:%s\n S:%s\nmessage:\n %s \ngot:\n  %s\nwanted:\n %s",
					algo, kat.Length, kat.N, kat.S, kat.Message, got, kat.Digest)
				t.Logf("wanted %+v", kat)
				t.FailNow()
			}
			continue
		}
	}
}*/

// TestKeccak does a basic test of the non-standardized Keccak hash functions.
func TestKeccak(t *testing.T) {
	tests := []struct {
		fn   func() hash.Hash
		data []byte
		want string
	}{

		{
			NewLegacyKeccak256,
			[]byte("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text"),
			"39670cbdfbfc25519e9834899e13569e5f5802b77df3d8259961f713ad09745d",
		},
		{
			NewLegacyKeccak256,
			[]byte("abc"),
			"4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45",
		},
		{
			NewLegacyKeccak256,
			[]byte("916889520661998333947422699691"),
			"c5b223cc231dd56389ca19435758fba38e79f3c461d989d818ceb96595d310d1",
		},
		{
			NewLegacyKeccak256,
			[]byte("6R1Q2KvQNGxI4OtUGDS1rSCbJXkp1H"),
			"8e94c6d49ca23597ee1d4a317b17a85ae38f6f3241e11c3ace4abe756287839a",
		},
		{
			NewLegacyKeccak256,
			[]byte("SOMERANDOMSTRING"),
			"59f47b8dc1ffdfb716d2fcb313c070040f6049bd23c321c6cb8adb3f79eb720f",
		},
		{
			NewLegacyKeccak256,
			[]byte("Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC"),
			"da5293ad50619d7fdac8b4c8dc2f80e0cd40b6571be14740229d0a1d6e2fc232",
		},
		{
			NewLegacyKeccak256,
			[]byte("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc."),
			"a84506d23b7246dc18805c8bf60316fb6f677934a546ee8d8650495929f21eef",
		},
	}

	for _, u := range tests {
		h := u.fn()
		h.Write(u.data)
		got := h.Sum(nil)
		want := decodeHex(u.want)
		if !bytes.Equal(got, want) {
			t.Errorf("unexpected hash for size %d: got '%x' want '%s'", h.Size()*8, got, u.want)
		}
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		data []byte
		want string
	}{

		{

			[]byte("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text"),
			"39670cbdfbfc25519e9834899e13569e5f5802b77df3d8259961f713ad09745d",
		},
		{

			[]byte("abc"),
			"4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45",
		},
		{

			[]byte("916889520661998333947422699691"),
			"c5b223cc231dd56389ca19435758fba38e79f3c461d989d818ceb96595d310d1",
		},
		{

			[]byte("6R1Q2KvQNGxI4OtUGDS1rSCbJXkp1H"),
			"8e94c6d49ca23597ee1d4a317b17a85ae38f6f3241e11c3ace4abe756287839a",
		},
		{

			[]byte("SOMERANDOMSTRING"),
			"59f47b8dc1ffdfb716d2fcb313c070040f6049bd23c321c6cb8adb3f79eb720f",
		},
		{

			[]byte("Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC"),
			"da5293ad50619d7fdac8b4c8dc2f80e0cd40b6571be14740229d0a1d6e2fc232",
		},
		{

			[]byte("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc."),
			"a84506d23b7246dc18805c8bf60316fb6f677934a546ee8d8650495929f21eef",
		},
	}

	for _, u := range tests {
		got := Hash(u.data)
		want := decodeHex(u.want)
		if !bytes.Equal(got[:], want) {
			t.Errorf("unexpected hash: got '%x' want '%s'", got, u.want)
		}
	}
}

// sequentialBytes produces a buffer of size consecutive bytes 0x00, 0x01, ..., used for testing.
//
// The alignment of each slice is intentionally randomized to detect alignment
// issues in the implementation. See https://golang.org/issue/37644.
// Ideally, the compiler should fuzz the alignment itself.
// (See https://golang.org/issue/35128.)
func sequentialBytes(size int) []byte {
	alignmentOffset := rand.IntN(8)
	result := make([]byte, size+alignmentOffset)[alignmentOffset:]
	for i := range result {
		result[i] = byte(i)
	}
	return result
}

/*func TestMarshalUnmarshal(t *testing.T) {
	t.Run("Keccak-256", func(t *testing.T) { testMarshalUnmarshal(t, NewLegacyKeccak256()) })
}*/

// TODO(filippo): move this to crypto/internal/cryptotest.
//XX: remove rand.Read
/*func testMarshalUnmarshal(t *testing.T, h hash.Hash) {
	buf := make([]byte, 200)
	rand.Read(buf)
	n := rand.IntN(200)
	h.Write(buf)
	want := h.Sum(nil)
	h.Reset()
	h.Write(buf[:n])
	b, err := h.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary: %v", err)
	}
	h.Write(bytes.Repeat([]byte{0}, 200))
	if err := h.(encoding.BinaryUnmarshaler).UnmarshalBinary(b); err != nil {
		t.Errorf("UnmarshalBinary: %v", err)
	}
	h.Write(buf[n:])
	got := h.Sum(nil)
	if !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}*/

// BenchmarkPermutationFunction measures the speed of the permutation function
// with no input data.
func BenchmarkPermutationFunction(b *testing.B) {
	b.SetBytes(int64(200))
	var lanes [25]uint64
	for i := 0; i < b.N; i++ {
		keccakF1600(&lanes)
	}
}

// benchmarkHash tests the speed to hash num buffers of buflen each.
func benchmarkHash(b *testing.B, h hash.Hash, size, num int) {
	b.StopTimer()
	h.Reset()
	data := sequentialBytes(size)
	b.SetBytes(int64(size * num))
	b.StartTimer()

	var state []byte
	for i := 0; i < b.N; i++ {
		for j := 0; j < num; j++ {
			h.Write(data)
		}
		state = h.Sum(state[:0])
	}
	b.StopTimer()
	h.Reset()
}
