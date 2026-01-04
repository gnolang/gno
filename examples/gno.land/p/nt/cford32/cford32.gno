// Modified from the Go Source code for encoding/base32.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cford32 implements a base32-like encoding/decoding package, with the
// encoding scheme [specified by Douglas Crockford].
//
// From the website, the requirements of said encoding scheme are to:
//
//   - Be human readable and machine readable.
//   - Be compact. Humans have difficulty in manipulating long strings of arbitrary symbols.
//   - Be error resistant. Entering the symbols must not require keyboarding gymnastics.
//   - Be pronounceable. Humans should be able to accurately transmit the symbols to other humans using a telephone.
//
// This is slightly different from a simple difference in encoding table from
// the Go's stdlib `encoding/base32`, as when decoding the characters i I l L are
// parsed as 1, and o O is parsed as 0.
//
// This package additionally provides ways to encode uint64's efficiently,
// as well as efficient encoding to a lowercase variation of the encoding.
// The encodings never use paddings.
//
// # Uint64 Encoding
//
// Aside from lower/uppercase encoding, there is a compact encoding, allowing
// to encode all values in [0,2^34), and the full encoding, allowing all
// values in [0,2^64). The compact encoding uses 7 characters, and the full
// encoding uses 13 characters. Both are parsed unambiguously by the Uint64
// decoder.
//
// The compact encodings have the first character between ['0','f'], while the
// full encoding's first character ranges between ['g','z']. Practically, in
// your usage of the package, you should consider which one to use and stick
// with it, while considering that the compact encoding, once it reaches 2^34,
// automatically switches to the full encoding. The properties of the generated
// strings are still maintained: for instance, any two encoded uint64s x,y
// consistently generated with the compact encoding, if the numeric value is
// x < y, will also be x < y in lexical ordering. However, values [0,2^34) have a
// "double encoding", which if mixed together lose the lexical ordering property.
//
// The Uint64 encoding is most useful for generating string versions of Uint64
// IDs. Practically, it allows you to retain sleek and compact IDs for your
// application for the first 2^34 (>17 billion) entities, while seamlessly
// rolling over to the full encoding should you exceed that. You are encouraged
// to use it unless you have a requirement or preferences for IDs consistently
// being always the same size.
//
// To use the cford32 encoding for IDs, you may want to consider using package
// [gno.land/p/nt/seqid].
//
// [specified by Douglas Crockford]: https://www.crockford.com/base32.html
package cford32

import (
	"io"
	"strconv"
)

const (
	encTable      = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	encTableLower = "0123456789abcdefghjkmnpqrstvwxyz"

	// each line is 16 bytes
	decTable = "" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" + // 00-0f
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" + // 10-1f
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" + // 20-2f
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" + // 30-3f
		"\xff\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x01\x12\x13\x01\x14\x15\x00" + // 40-4f
		"\x16\x17\x18\x19\x1a\xff\x1b\x1c\x1d\x1e\x1f\xff\xff\xff\xff\xff" + // 50-5f
		"\xff\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x01\x12\x13\x01\x14\x15\x00" + // 60-6f
		"\x16\x17\x18\x19\x1a\xff\x1b\x1c\x1d\x1e\x1f\xff\xff\xff\xff\xff" + // 70-7f
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" + // 80-ff (not ASCII)
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"
)

// CorruptInputError is returned by parsing functions when an invalid character
// in the input is found. The integer value represents the byte index where
// the error occurred.
//
// This is typically because the given character does not exist in the encoding.
type CorruptInputError int64

func (e CorruptInputError) Error() string {
	return "illegal cford32 data at input byte " + strconv.FormatInt(int64(e), 10)
}

// Uint64 parses a cford32-encoded byte slice into a uint64.
//
//   - The parser requires all provided character to be valid cford32 characters.
//   - The parser disregards case.
//   - If the first character is '0' <= c <= 'f', then the passed value is assumed
//     encoded in the compact encoding, and must be 7 characters long.
//   - If the first character is 'g' <= c <= 'z',  then the passed value is
//     assumed encoded in the full encoding, and must be 13 characters long.
//
// If any of these requirements fail, a CorruptInputError will be returned.
func Uint64(b []byte) (uint64, error) {
	if len(b) == 0 {
		return 0, CorruptInputError(0)
	}
	b0 := decTable[b[0]]
	switch {
	default:
		return 0, CorruptInputError(0)
	case len(b) == 7 && b0 < 16:
		decVals := [7]byte{
			decTable[b[0]],
			decTable[b[1]],
			decTable[b[2]],
			decTable[b[3]],
			decTable[b[4]],
			decTable[b[5]],
			decTable[b[6]],
		}
		for idx, v := range decVals {
			if v >= 32 {
				return 0, CorruptInputError(idx)
			}
		}

		return 0 +
			uint64(decVals[0])<<30 |
			uint64(decVals[1])<<25 |
			uint64(decVals[2])<<20 |
			uint64(decVals[3])<<15 |
			uint64(decVals[4])<<10 |
			uint64(decVals[5])<<5 |
			uint64(decVals[6]), nil
	case len(b) == 13 && b0 >= 16 && b0 < 32:
		decVals := [13]byte{
			decTable[b[0]] & 0x0F, // disregard high bit
			decTable[b[1]],
			decTable[b[2]],
			decTable[b[3]],
			decTable[b[4]],
			decTable[b[5]],
			decTable[b[6]],
			decTable[b[7]],
			decTable[b[8]],
			decTable[b[9]],
			decTable[b[10]],
			decTable[b[11]],
			decTable[b[12]],
		}
		for idx, v := range decVals {
			if v >= 32 {
				return 0, CorruptInputError(idx)
			}
		}

		return 0 +
			uint64(decVals[0])<<60 |
			uint64(decVals[1])<<55 |
			uint64(decVals[2])<<50 |
			uint64(decVals[3])<<45 |
			uint64(decVals[4])<<40 |
			uint64(decVals[5])<<35 |
			uint64(decVals[6])<<30 |
			uint64(decVals[7])<<25 |
			uint64(decVals[8])<<20 |
			uint64(decVals[9])<<15 |
			uint64(decVals[10])<<10 |
			uint64(decVals[11])<<5 |
			uint64(decVals[12]), nil
	}
}

const mask = 31

// PutUint64 returns a cford32-encoded byte slice.
func PutUint64(id uint64) [13]byte {
	return [13]byte{
		encTable[id>>60&mask|0x10], // specify full encoding
		encTable[id>>55&mask],
		encTable[id>>50&mask],
		encTable[id>>45&mask],
		encTable[id>>40&mask],
		encTable[id>>35&mask],
		encTable[id>>30&mask],
		encTable[id>>25&mask],
		encTable[id>>20&mask],
		encTable[id>>15&mask],
		encTable[id>>10&mask],
		encTable[id>>5&mask],
		encTable[id&mask],
	}
}

// PutUint64Lower returns a cford32-encoded byte array, swapping uppercase
// letters with lowercase.
//
// For more information on how the value is encoded, see [Uint64].
func PutUint64Lower(id uint64) [13]byte {
	return [13]byte{
		encTableLower[id>>60&mask|0x10],
		encTableLower[id>>55&mask],
		encTableLower[id>>50&mask],
		encTableLower[id>>45&mask],
		encTableLower[id>>40&mask],
		encTableLower[id>>35&mask],
		encTableLower[id>>30&mask],
		encTableLower[id>>25&mask],
		encTableLower[id>>20&mask],
		encTableLower[id>>15&mask],
		encTableLower[id>>10&mask],
		encTableLower[id>>5&mask],
		encTableLower[id&mask],
	}
}

// PutCompact returns a cford32-encoded byte slice, using the compact
// representation of cford32 described in the package documentation where
// possible (all values of id < 1<<34). The lowercase encoding is used.
//
// The resulting byte slice will be 7 bytes long for all compact values,
// and 13 bytes long for
func PutCompact(id uint64) []byte {
	return AppendCompact(id, nil)
}

// AppendCompact works like [PutCompact] but appends to the given byte slice
// instead of allocating one anew.
func AppendCompact(id uint64, b []byte) []byte {
	const maxCompact = 1 << 34
	if id < maxCompact {
		return append(b,
			encTableLower[id>>30&mask],
			encTableLower[id>>25&mask],
			encTableLower[id>>20&mask],
			encTableLower[id>>15&mask],
			encTableLower[id>>10&mask],
			encTableLower[id>>5&mask],
			encTableLower[id&mask],
		)
	}
	return append(b,
		encTableLower[id>>60&mask|0x10],
		encTableLower[id>>55&mask],
		encTableLower[id>>50&mask],
		encTableLower[id>>45&mask],
		encTableLower[id>>40&mask],
		encTableLower[id>>35&mask],
		encTableLower[id>>30&mask],
		encTableLower[id>>25&mask],
		encTableLower[id>>20&mask],
		encTableLower[id>>15&mask],
		encTableLower[id>>10&mask],
		encTableLower[id>>5&mask],
		encTableLower[id&mask],
	)
}

func DecodedLen(n int) int {
	return n/8*5 + n%8*5/8
}

func EncodedLen(n int) int {
	return n/5*8 + (n%5*8+4)/5
}

// Encode encodes src using the encoding enc,
// writing [EncodedLen](len(src)) bytes to dst.
//
// The encoding does not contain any padding, unlike Go's base32.
func Encode(dst, src []byte) {
	// Copied from encoding/base32/base32.go (go1.22)
	if len(src) == 0 {
		return
	}

	di, si := 0, 0
	n := (len(src) / 5) * 5
	for si < n {
		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		hi := uint32(src[si+0])<<24 | uint32(src[si+1])<<16 | uint32(src[si+2])<<8 | uint32(src[si+3])
		lo := hi<<8 | uint32(src[si+4])

		dst[di+0] = encTable[(hi>>27)&0x1F]
		dst[di+1] = encTable[(hi>>22)&0x1F]
		dst[di+2] = encTable[(hi>>17)&0x1F]
		dst[di+3] = encTable[(hi>>12)&0x1F]
		dst[di+4] = encTable[(hi>>7)&0x1F]
		dst[di+5] = encTable[(hi>>2)&0x1F]
		dst[di+6] = encTable[(lo>>5)&0x1F]
		dst[di+7] = encTable[(lo)&0x1F]

		si += 5
		di += 8
	}

	// Add the remaining small block
	remain := len(src) - si
	if remain == 0 {
		return
	}

	// Encode the remaining bytes in reverse order.
	val := uint32(0)
	switch remain {
	case 4:
		val |= uint32(src[si+3])
		dst[di+6] = encTable[val<<3&0x1F]
		dst[di+5] = encTable[val>>2&0x1F]
		fallthrough
	case 3:
		val |= uint32(src[si+2]) << 8
		dst[di+4] = encTable[val>>7&0x1F]
		fallthrough
	case 2:
		val |= uint32(src[si+1]) << 16
		dst[di+3] = encTable[val>>12&0x1F]
		dst[di+2] = encTable[val>>17&0x1F]
		fallthrough
	case 1:
		val |= uint32(src[si+0]) << 24
		dst[di+1] = encTable[val>>22&0x1F]
		dst[di+0] = encTable[val>>27&0x1F]
	}
}

// EncodeLower is like [Encode], but uses the lowercase
func EncodeLower(dst, src []byte) {
	// Copied from encoding/base32/base32.go (go1.22)
	if len(src) == 0 {
		return
	}

	di, si := 0, 0
	n := (len(src) / 5) * 5
	for si < n {
		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		hi := uint32(src[si+0])<<24 | uint32(src[si+1])<<16 | uint32(src[si+2])<<8 | uint32(src[si+3])
		lo := hi<<8 | uint32(src[si+4])

		dst[di+0] = encTableLower[(hi>>27)&0x1F]
		dst[di+1] = encTableLower[(hi>>22)&0x1F]
		dst[di+2] = encTableLower[(hi>>17)&0x1F]
		dst[di+3] = encTableLower[(hi>>12)&0x1F]
		dst[di+4] = encTableLower[(hi>>7)&0x1F]
		dst[di+5] = encTableLower[(hi>>2)&0x1F]
		dst[di+6] = encTableLower[(lo>>5)&0x1F]
		dst[di+7] = encTableLower[(lo)&0x1F]

		si += 5
		di += 8
	}

	// Add the remaining small block
	remain := len(src) - si
	if remain == 0 {
		return
	}

	// Encode the remaining bytes in reverse order.
	val := uint32(0)
	switch remain {
	case 4:
		val |= uint32(src[si+3])
		dst[di+6] = encTableLower[val<<3&0x1F]
		dst[di+5] = encTableLower[val>>2&0x1F]
		fallthrough
	case 3:
		val |= uint32(src[si+2]) << 8
		dst[di+4] = encTableLower[val>>7&0x1F]
		fallthrough
	case 2:
		val |= uint32(src[si+1]) << 16
		dst[di+3] = encTableLower[val>>12&0x1F]
		dst[di+2] = encTableLower[val>>17&0x1F]
		fallthrough
	case 1:
		val |= uint32(src[si+0]) << 24
		dst[di+1] = encTableLower[val>>22&0x1F]
		dst[di+0] = encTableLower[val>>27&0x1F]
	}
}

// AppendEncode appends the cford32 encoded src to dst
// and returns the extended buffer.
func AppendEncode(dst, src []byte) []byte {
	n := EncodedLen(len(src))
	dst = grow(dst, n)
	Encode(dst[len(dst):][:n], src)
	return dst[:len(dst)+n]
}

// AppendEncodeLower appends the lowercase cford32 encoded src to dst
// and returns the extended buffer.
func AppendEncodeLower(dst, src []byte) []byte {
	n := EncodedLen(len(src))
	dst = grow(dst, n)
	EncodeLower(dst[len(dst):][:n], src)
	return dst[:len(dst)+n]
}

func grow(s []byte, n int) []byte {
	// slices.Grow
	if n -= cap(s) - len(s); n > 0 {
		news := make([]byte, cap(s)+n)
		copy(news[:cap(s)], s[:cap(s)])
		return news[:len(s)]
	}
	return s
}

// EncodeToString returns the cford32 encoding of src.
func EncodeToString(src []byte) string {
	buf := make([]byte, EncodedLen(len(src)))
	Encode(buf, src)
	return string(buf)
}

// EncodeToStringLower returns the cford32 lowercase encoding of src.
func EncodeToStringLower(src []byte) string {
	buf := make([]byte, EncodedLen(len(src)))
	EncodeLower(buf, src)
	return string(buf)
}

func decode(dst, src []byte) (n int, err error) {
	dsti := 0
	olen := len(src)

	for len(src) > 0 {
		// Decode quantum using the base32 alphabet
		var dbuf [8]byte
		dlen := 8

		for j := 0; j < 8; {
			if len(src) == 0 {
				// We have reached the end and are not expecting any padding
				dlen = j
				break
			}
			in := src[0]
			src = src[1:]
			dbuf[j] = decTable[in]
			if dbuf[j] == 0xFF {
				return n, CorruptInputError(olen - len(src) - 1)
			}
			j++
		}

		// Pack 8x 5-bit source blocks into 5 byte destination
		// quantum
		switch dlen {
		case 8:
			dst[dsti+4] = dbuf[6]<<5 | dbuf[7]
			n++
			fallthrough
		case 7:
			dst[dsti+3] = dbuf[4]<<7 | dbuf[5]<<2 | dbuf[6]>>3
			n++
			fallthrough
		case 5:
			dst[dsti+2] = dbuf[3]<<4 | dbuf[4]>>1
			n++
			fallthrough
		case 4:
			dst[dsti+1] = dbuf[1]<<6 | dbuf[2]<<1 | dbuf[3]>>4
			n++
			fallthrough
		case 2:
			dst[dsti+0] = dbuf[0]<<3 | dbuf[1]>>2
			n++
		}
		dsti += 5
	}
	return n, nil
}

type encoder struct {
	err  error
	w    io.Writer
	enc  func(dst, src []byte)
	buf  [5]byte    // buffered data waiting to be encoded
	nbuf int        // number of bytes in buf
	out  [1024]byte // output buffer
}

func NewEncoder(w io.Writer) io.WriteCloser {
	return &encoder{w: w, enc: Encode}
}

func NewEncoderLower(w io.Writer) io.WriteCloser {
	return &encoder{w: w, enc: EncodeLower}
}

func (e *encoder) Write(p []byte) (n int, err error) {
	if e.err != nil {
		return 0, e.err
	}

	// Leading fringe.
	if e.nbuf > 0 {
		var i int
		for i = 0; i < len(p) && e.nbuf < 5; i++ {
			e.buf[e.nbuf] = p[i]
			e.nbuf++
		}
		n += i
		p = p[i:]
		if e.nbuf < 5 {
			return
		}
		e.enc(e.out[0:], e.buf[0:])
		if _, e.err = e.w.Write(e.out[0:8]); e.err != nil {
			return n, e.err
		}
		e.nbuf = 0
	}

	// Large interior chunks.
	for len(p) >= 5 {
		nn := len(e.out) / 8 * 5
		if nn > len(p) {
			nn = len(p)
			nn -= nn % 5
		}
		e.enc(e.out[0:], p[0:nn])
		if _, e.err = e.w.Write(e.out[0 : nn/5*8]); e.err != nil {
			return n, e.err
		}
		n += nn
		p = p[nn:]
	}

	// Trailing fringe.
	copy(e.buf[:], p)
	e.nbuf = len(p)
	n += len(p)
	return
}

// Close flushes any pending output from the encoder.
// It is an error to call Write after calling Close.
func (e *encoder) Close() error {
	// If there's anything left in the buffer, flush it out
	if e.err == nil && e.nbuf > 0 {
		e.enc(e.out[0:], e.buf[0:e.nbuf])
		encodedLen := EncodedLen(e.nbuf)
		e.nbuf = 0
		_, e.err = e.w.Write(e.out[0:encodedLen])
	}
	return e.err
}

// Decode decodes src using cford32. It writes at most
// [DecodedLen](len(src)) bytes to dst and returns the number of bytes
// written. If src contains invalid cford32 data, it will return the
// number of bytes successfully written and [CorruptInputError].
// Newline characters (\r and \n) are ignored.
func Decode(dst, src []byte) (n int, err error) {
	buf := make([]byte, len(src))
	l := stripNewlines(buf, src)
	return decode(dst, buf[:l])
}

// AppendDecode appends the cford32 decoded src to dst
// and returns the extended buffer.
// If the input is malformed, it returns the partially decoded src and an error.
func AppendDecode(dst, src []byte) ([]byte, error) {
	n := DecodedLen(len(src))

	dst = grow(dst, n)
	dstsl := dst[len(dst) : len(dst)+n]
	n, err := Decode(dstsl, src)
	return dst[:len(dst)+n], err
}

// DecodeString returns the bytes represented by the cford32 string s.
func DecodeString(s string) ([]byte, error) {
	buf := []byte(s)
	l := stripNewlines(buf, buf)
	n, err := decode(buf, buf[:l])
	return buf[:n], err
}

// stripNewlines removes newline characters and returns the number
// of non-newline characters copied to dst.
func stripNewlines(dst, src []byte) int {
	offset := 0
	for _, b := range src {
		if b == '\r' || b == '\n' {
			continue
		}
		dst[offset] = b
		offset++
	}
	return offset
}

type decoder struct {
	err    error
	r      io.Reader
	buf    [1024]byte // leftover input
	nbuf   int
	out    []byte // leftover decoded output
	outbuf [1024 / 8 * 5]byte
}

// NewDecoder constructs a new base32 stream decoder.
func NewDecoder(r io.Reader) io.Reader {
	return &decoder{r: &newlineFilteringReader{r}}
}

func readEncodedData(r io.Reader, buf []byte) (n int, err error) {
	for n < 1 && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	return
}

func (d *decoder) Read(p []byte) (n int, err error) {
	// Use leftover decoded output from last read.
	if len(d.out) > 0 {
		n = copy(p, d.out)
		d.out = d.out[n:]
		if len(d.out) == 0 {
			return n, d.err
		}
		return n, nil
	}

	if d.err != nil {
		return 0, d.err
	}

	// Read nn bytes from input, bounded [8,len(d.buf)]
	nn := (len(p)/5 + 1) * 8
	if nn > len(d.buf) {
		nn = len(d.buf)
	}

	nn, d.err = readEncodedData(d.r, d.buf[d.nbuf:nn])
	d.nbuf += nn
	if d.nbuf < 1 {
		return 0, d.err
	}

	// Decode chunk into p, or d.out and then p if p is too small.
	nr := d.nbuf
	if d.err != io.EOF && nr%8 != 0 {
		nr -= nr % 8
	}
	nw := DecodedLen(d.nbuf)

	if nw > len(p) {
		nw, err = decode(d.outbuf[0:], d.buf[0:nr])
		d.out = d.outbuf[0:nw]
		n = copy(p, d.out)
		d.out = d.out[n:]
	} else {
		n, err = decode(p, d.buf[0:nr])
	}
	d.nbuf -= nr
	for i := 0; i < d.nbuf; i++ {
		d.buf[i] = d.buf[i+nr]
	}

	if err != nil && (d.err == nil || d.err == io.EOF) {
		d.err = err
	}

	if len(d.out) > 0 {
		// We cannot return all the decoded bytes to the caller in this
		// invocation of Read, so we return a nil error to ensure that Read
		// will be called again.  The error stored in d.err, if any, will be
		// returned with the last set of decoded bytes.
		return n, nil
	}

	return n, d.err
}

type newlineFilteringReader struct {
	wrapped io.Reader
}

func (r *newlineFilteringReader) Read(p []byte) (int, error) {
	n, err := r.wrapped.Read(p)
	for n > 0 {
		s := p[0:n]
		offset := stripNewlines(s, s)
		if err != nil || offset > 0 {
			return offset, err
		}
		// Previous buffer entirely whitespace, read again
		n, err = r.wrapped.Read(p)
	}
	return n, err
}
