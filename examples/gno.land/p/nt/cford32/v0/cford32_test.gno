package cford32

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestCompactRoundtrip(t *testing.T) {
	buf := make([]byte, 13)
	prev := make([]byte, 13)
	for i := uint64(0); i < (1 << 10); i++ {
		res := AppendCompact(i, buf[:0])
		back, err := Uint64(res)
		testEqual(t, "Uint64(%q) = (%d, %v), want %v", string(res), back, err, nil)
		testEqual(t, "Uint64(%q) = %d, want %v", string(res), back, i)

		testEqual(t, "bytes.Compare(prev, res) = %d, want %d", bytes.Compare(prev, res), -1)
		prev, buf = res, prev
	}
	for i := uint64(1<<34 - 1024); i < (1<<34 + 1024); i++ {
		res := AppendCompact(i, buf[:0])
		back, err := Uint64(res)
		// println(string(res))
		testEqual(t, "Uint64(%q) = (%d, %v), want %v", string(res), back, err, nil)
		testEqual(t, "Uint64(%q) = %d, want %v", string(res), back, i)

		testEqual(t, "bytes.Compare(prev, res) = %d, want %d", bytes.Compare(prev, res), -1)
		prev, buf = res, prev
	}
	for i := uint64(1<<64 - 5000); i != 0; i++ {
		res := AppendCompact(i, buf[:0])
		back, err := Uint64(res)
		testEqual(t, "Uint64(%q) = (%d, %v), want %v", string(res), back, err, nil)
		testEqual(t, "Uint64(%q) = %d, want %v", string(res), back, i)

		testEqual(t, "bytes.Compare(prev, res) = %d, want %d", bytes.Compare(prev, res), -1)
		prev, buf = res, prev
	}
}

func BenchmarkCompact(b *testing.B) {
	buf := make([]byte, 13)
	for i := 0; i < b.N; i++ {
		_ = AppendCompact(uint64(i), buf[:0])
	}
}

func TestUint64(t *testing.T) {
	tt := []struct {
		val    string
		output uint64
		err    string
	}{
		{"0000001", 1, ""},
		{"OoOoOoL", 1, ""},
		{"OoUoOoL", 0, CorruptInputError(2).Error()},
		{"!123123", 0, CorruptInputError(0).Error()},
		{"Loooooo", 1073741824, ""},
		{"goooooo", 0, CorruptInputError(0).Error()},
		{"goooooooooooo", 0, ""},
		{"goooooooooolo", 32, ""},
		{"fzzzzzz", (1 << 34) - 1, ""},
		{"g00000fzzzzzz", (1 << 34) - 1, ""},
		{"g000000", 0, CorruptInputError(0).Error()},
		{"g00000g000000", (1 << 34), ""},
	}

	for _, tc := range tt {
		t.Run(tc.val, func(t *testing.T) {
			res, err := Uint64([]byte(tc.val))
			if tc.err != "" {
				_ = uassert.Error(t, err) &&
					uassert.Equal(t, tc.err, err.Error())
			} else {
				_ = uassert.NoError(t, err) &&
					uassert.Equal(t, tc.output, res)
			}
		})
	}
}

func TestRandomCompactRoundtrip(t *testing.T) {
	for i := 0; i < 1<<12; i++ {
		value := rand.Uint64()
		encoded := PutCompact(value)
		decoded, err := Uint64(encoded)
		uassert.NoError(t, err)
		uassert.Equal(t, value, decoded)
	}
}

type testpair struct {
	decoded, encoded string
}

var pairs = []testpair{
	{"", ""},
	{"f", "CR"},
	{"fo", "CSQG"},
	{"foo", "CSQPY"},
	{"foob", "CSQPYRG"},
	{"fooba", "CSQPYRK1"},
	{"foobar", "CSQPYRK1E8"},

	{"sure.", "EDTQ4S9E"},
	{"sure", "EDTQ4S8"},
	{"sur", "EDTQ4"},
	{"su", "EDTG"},
	{"leasure.", "DHJP2WVNE9JJW"},
	{"easure.", "CNGQ6XBJCMQ0"},
	{"asure.", "C5SQAWK55R"},
}

var bigtest = testpair{
	"Twas brillig, and the slithy toves",
	"AHVP2WS0C9S6JV3CD5KJR831DSJ20X38CMG76V39EHM7J83MDXV6AWR",
}

func testEqual(t *testing.T, msg string, args ...any) bool {
	t.Helper()
	if args[len(args)-2] != args[len(args)-1] {
		t.Errorf(msg, args...)
		return false
	}
	return true
}

func TestEncode(t *testing.T) {
	for _, p := range pairs {
		got := EncodeToString([]byte(p.decoded))
		testEqual(t, "Encode(%q) = %q, want %q", p.decoded, got, p.encoded)
		dst := AppendEncode([]byte("lead"), []byte(p.decoded))
		testEqual(t, `AppendEncode("lead", %q) = %q, want %q`, p.decoded, string(dst), "lead"+p.encoded)
	}
}

func TestEncoder(t *testing.T) {
	for _, p := range pairs {
		bb := &strings.Builder{}
		encoder := NewEncoder(bb)
		encoder.Write([]byte(p.decoded))
		encoder.Close()
		testEqual(t, "Encode(%q) = %q, want %q", p.decoded, bb.String(), p.encoded)
	}
}

func TestEncoderBuffering(t *testing.T) {
	input := []byte(bigtest.decoded)
	for bs := 1; bs <= 12; bs++ {
		bb := &strings.Builder{}
		encoder := NewEncoder(bb)
		for pos := 0; pos < len(input); pos += bs {
			end := pos + bs
			if end > len(input) {
				end = len(input)
			}
			n, err := encoder.Write(input[pos:end])
			testEqual(t, "Write(%q) gave error %v, want %v", input[pos:end], err, error(nil))
			testEqual(t, "Write(%q) gave length %v, want %v", input[pos:end], n, end-pos)
		}
		err := encoder.Close()
		testEqual(t, "Close gave error %v, want %v", err, error(nil))
		testEqual(t, "Encoding/%d of %q = %q, want %q", bs, bigtest.decoded, bb.String(), bigtest.encoded)
	}
}

func TestDecode(t *testing.T) {
	for _, p := range pairs {
		dbuf := make([]byte, DecodedLen(len(p.encoded)))
		count, err := decode(dbuf, []byte(p.encoded))
		testEqual(t, "Decode(%q) = error %v, want %v", p.encoded, err, error(nil))
		testEqual(t, "Decode(%q) = length %v, want %v", p.encoded, count, len(p.decoded))
		testEqual(t, "Decode(%q) = %q, want %q", p.encoded, string(dbuf[0:count]), p.decoded)

		dbuf, err = DecodeString(p.encoded)
		testEqual(t, "DecodeString(%q) = error %v, want %v", p.encoded, err, error(nil))
		testEqual(t, "DecodeString(%q) = %q, want %q", p.encoded, string(dbuf), p.decoded)

		// XXX: https://github.com/gnolang/gno/issues/1570
		dst, err := AppendDecode(append([]byte(nil), []byte("lead")...), []byte(p.encoded))
		testEqual(t, "AppendDecode(%q) = error %v, want %v", p.encoded, err, error(nil))
		testEqual(t, `AppendDecode("lead", %q) = %q, want %q`, p.encoded, string(dst), "lead"+p.decoded)

		dst2, err := AppendDecode(dst[:0:len(p.decoded)], []byte(p.encoded))
		testEqual(t, "AppendDecode(%q) = error %v, want %v", p.encoded, err, error(nil))
		testEqual(t, `AppendDecode("", %q) = %q, want %q`, p.encoded, string(dst2), p.decoded)
		// XXX: https://github.com/gnolang/gno/issues/1569
		// old used &dst2[0] != &dst[0] as a check.
		if len(dst) > 0 && len(dst2) > 0 && cap(dst2) != len(p.decoded) {
			t.Errorf("unexpected capacity growth: got %d, want %d", cap(dst2), len(p.decoded))
		}
	}
}

// A minimal variation on strings.Reader.
// Here, we return a io.EOF immediately on Read if the read has reached the end
// of the reader. It's used to simplify TestDecoder.
type stringReader struct {
	s string
	i int64
}

func (r *stringReader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	if r.i >= int64(len(r.s)) {
		return n, io.EOF
	}
	return
}

func TestDecoder(t *testing.T) {
	for _, p := range pairs {
		decoder := NewDecoder(&stringReader{p.encoded, 0})
		dbuf := make([]byte, DecodedLen(len(p.encoded)))
		count, err := decoder.Read(dbuf)
		if err != nil && err != io.EOF {
			t.Fatal("Read failed", err)
		}
		testEqual(t, "Read from %q = length %v, want %v", p.encoded, count, len(p.decoded))
		testEqual(t, "Decoding of %q = %q, want %q", p.encoded, string(dbuf[0:count]), p.decoded)
		if err != io.EOF {
			_, err = decoder.Read(dbuf)
		}
		testEqual(t, "Read from %q = %v, want %v", p.encoded, err, io.EOF)
	}
}

type badReader struct {
	data   []byte
	errs   []error
	called int
	limit  int
}

// Populates p with data, returns a count of the bytes written and an
// error.  The error returned is taken from badReader.errs, with each
// invocation of Read returning the next error in this slice, or io.EOF,
// if all errors from the slice have already been returned.  The
// number of bytes returned is determined by the size of the input buffer
// the test passes to decoder.Read and will be a multiple of 8, unless
// badReader.limit is non zero.
func (b *badReader) Read(p []byte) (int, error) {
	lim := len(p)
	if b.limit != 0 && b.limit < lim {
		lim = b.limit
	}
	if len(b.data) < lim {
		lim = len(b.data)
	}
	for i := range p[:lim] {
		p[i] = b.data[i]
	}
	b.data = b.data[lim:]
	err := io.EOF
	if b.called < len(b.errs) {
		err = b.errs[b.called]
	}
	b.called++
	return lim, err
}

// TestIssue20044 tests that decoder.Read behaves correctly when the caller
// supplied reader returns an error.
func TestIssue20044(t *testing.T) {
	badErr := errors.New("bad reader error")
	testCases := []struct {
		r       badReader
		res     string
		err     error
		dbuflen int
	}{
		// Check valid input data accompanied by an error is processed and the error is propagated.
		{
			r:   badReader{data: []byte("d1jprv3fexqq4v34"), errs: []error{badErr}},
			res: "helloworld", err: badErr,
		},
		// Check a read error accompanied by input data consisting of newlines only is propagated.
		{
			r:   badReader{data: []byte("\n\n\n\n\n\n\n\n"), errs: []error{badErr, nil}},
			res: "", err: badErr,
		},
		// Reader will be called twice.  The first time it will return 8 newline characters.  The
		// second time valid base32 encoded data and an error.  The data should be decoded
		// correctly and the error should be propagated.
		{
			r:   badReader{data: []byte("\n\n\n\n\n\n\n\nd1jprv3fexqq4v34"), errs: []error{nil, badErr}},
			res: "helloworld", err: badErr, dbuflen: 8,
		},
		// Reader returns invalid input data (too short) and an error.  Verify the reader
		// error is returned.
		{
			r:   badReader{data: []byte("c"), errs: []error{badErr}},
			res: "", err: badErr,
		},
		// Reader returns invalid input data (too short) but no error.  Verify io.ErrUnexpectedEOF
		// is returned.
		// NOTE(thehowl): I don't think this should applyto us?
		/* {
			r:   badReader{data: []byte("c"), errs: []error{nil}},
			res: "", err: io.ErrUnexpectedEOF,
		},*/
		// Reader returns invalid input data and an error.  Verify the reader and not the
		// decoder error is returned.
		{
			r:   badReader{data: []byte("cu"), errs: []error{badErr}},
			res: "", err: badErr,
		},
		// Reader returns valid data and io.EOF.  Check data is decoded and io.EOF is propagated.
		{
			r:   badReader{data: []byte("csqpyrk1"), errs: []error{io.EOF}},
			res: "fooba", err: io.EOF,
		},
		// Check errors are properly reported when decoder.Read is called multiple times.
		// decoder.Read will be called 8 times, badReader.Read will be called twice, returning
		// valid data both times but an error on the second call.
		{
			r:   badReader{data: []byte("dhjp2wvne9jjwc9g"), errs: []error{nil, badErr}},
			res: "leasure.10", err: badErr, dbuflen: 1,
		},
		// Check io.EOF is properly reported when decoder.Read is called multiple times.
		// decoder.Read will be called 8 times, badReader.Read will be called twice, returning
		// valid data both times but io.EOF on the second call.
		{
			r:   badReader{data: []byte("dhjp2wvne9jjw"), errs: []error{nil, io.EOF}},
			res: "leasure.", err: io.EOF, dbuflen: 1,
		},
		// The following two test cases check that errors are propagated correctly when more than
		// 8 bytes are read at a time.
		{
			r:   badReader{data: []byte("dhjp2wvne9jjw"), errs: []error{io.EOF}},
			res: "leasure.", err: io.EOF, dbuflen: 11,
		},
		{
			r:   badReader{data: []byte("dhjp2wvne9jjwc9g"), errs: []error{badErr}},
			res: "leasure.10", err: badErr, dbuflen: 11,
		},
		// Check that errors are correctly propagated when the reader returns valid bytes in
		// groups that are not divisible by 8.  The first read will return 11 bytes and no
		// error.  The second will return 7 and an error.  The data should be decoded correctly
		// and the error should be propagated.
		// NOTE(thehowl): again, this is on the assumption that this is padded, and it's not.
		/* {
			r:   badReader{data: []byte("dhjp2wvne9jjw"), errs: []error{nil, badErr}, limit: 11},
			res: "leasure.", err: badErr,
		}, */
	}

	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", idx, string(tc.res)), func(t *testing.T) {
			input := tc.r.data
			decoder := NewDecoder(&tc.r)
			var dbuflen int
			if tc.dbuflen > 0 {
				dbuflen = tc.dbuflen
			} else {
				dbuflen = DecodedLen(len(input))
			}
			dbuf := make([]byte, dbuflen)
			var err error
			var res []byte
			for err == nil {
				var n int
				n, err = decoder.Read(dbuf)
				if n > 0 {
					res = append(res, dbuf[:n]...)
				}
			}

			testEqual(t, "Decoding of %q = %q, want %q", string(input), string(res), tc.res)
			testEqual(t, "Decoding of %q err = %v, expected %v", string(input), err, tc.err)
		})
	}
}

// TestDecoderError verifies decode errors are propagated when there are no read
// errors.
func TestDecoderError(t *testing.T) {
	for _, readErr := range []error{io.EOF, nil} {
		input := "ucsqpyrk1u"
		dbuf := make([]byte, DecodedLen(len(input)))
		br := badReader{data: []byte(input), errs: []error{readErr}}
		decoder := NewDecoder(&br)
		n, err := decoder.Read(dbuf)
		testEqual(t, "Read after EOF, n = %d, expected %d", n, 0)
		if _, ok := err.(CorruptInputError); !ok {
			t.Errorf("Corrupt input error expected.  Found %T", err)
		}
	}
}

// TestReaderEOF ensures decoder.Read behaves correctly when input data is
// exhausted.
func TestReaderEOF(t *testing.T) {
	for _, readErr := range []error{io.EOF, nil} {
		input := "MZXW6YTB"
		br := badReader{data: []byte(input), errs: []error{nil, readErr}}
		decoder := NewDecoder(&br)
		dbuf := make([]byte, DecodedLen(len(input)))
		n, err := decoder.Read(dbuf)
		testEqual(t, "Decoding of %q err = %v, expected %v", input, err, error(nil))
		n, err = decoder.Read(dbuf)
		testEqual(t, "Read after EOF, n = %d, expected %d", n, 0)
		testEqual(t, "Read after EOF, err = %v, expected %v", err, io.EOF)
		n, err = decoder.Read(dbuf)
		testEqual(t, "Read after EOF, n = %d, expected %d", n, 0)
		testEqual(t, "Read after EOF, err = %v, expected %v", err, io.EOF)
	}
}

func TestDecoderBuffering(t *testing.T) {
	for bs := 1; bs <= 12; bs++ {
		decoder := NewDecoder(strings.NewReader(bigtest.encoded))
		buf := make([]byte, len(bigtest.decoded)+12)
		var total int
		var n int
		var err error
		for total = 0; total < len(bigtest.decoded) && err == nil; {
			n, err = decoder.Read(buf[total : total+bs])
			total += n
		}
		if err != nil && err != io.EOF {
			t.Errorf("Read from %q at pos %d = %d, unexpected error %v", bigtest.encoded, total, n, err)
		}
		testEqual(t, "Decoding/%d of %q = %q, want %q", bs, bigtest.encoded, string(buf[0:total]), bigtest.decoded)
	}
}

func TestDecodeCorrupt(t *testing.T) {
	testCases := []struct {
		input  string
		offset int // -1 means no corruption.
	}{
		{"", -1},
		{"iIoOlL", -1},
		{"!!!!", 0},
		{"uxp10", 0},
		{"x===", 1},
		{"AA=A====", 2},
		{"AAA=AAAA", 3},
		// Much fewer cases compared to Go as there are much fewer cases where input
		// can be "corrupted".
	}
	for _, tc := range testCases {
		dbuf := make([]byte, DecodedLen(len(tc.input)))
		_, err := Decode(dbuf, []byte(tc.input))
		if tc.offset == -1 {
			if err != nil {
				t.Error("Decoder wrongly detected corruption in", tc.input)
			}
			continue
		}
		switch err := err.(type) {
		case CorruptInputError:
			testEqual(t, "Corruption in %q at offset %v, want %v", tc.input, int(err), tc.offset)
		default:
			t.Error("Decoder failed to detect corruption in", tc)
		}
	}
}

func TestBig(t *testing.T) {
	n := 3*1000 + 1
	raw := make([]byte, n)
	const alpha = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < n; i++ {
		raw[i] = alpha[i%len(alpha)]
	}
	encoded := new(bytes.Buffer)
	w := NewEncoder(encoded)
	nn, err := w.Write(raw)
	if nn != n || err != nil {
		t.Fatalf("Encoder.Write(raw) = %d, %v want %d, nil", nn, err, n)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("Encoder.Close() = %v want nil", err)
	}
	decoded, err := io.ReadAll(NewDecoder(encoded))
	if err != nil {
		t.Fatalf("io.ReadAll(NewDecoder(...)): %v", err)
	}

	if !bytes.Equal(raw, decoded) {
		var i int
		for i = 0; i < len(decoded) && i < len(raw); i++ {
			if decoded[i] != raw[i] {
				break
			}
		}
		t.Errorf("Decode(Encode(%d-byte string)) failed at offset %d", n, i)
	}
}

func testStringEncoding(t *testing.T, expected string, examples []string) {
	for _, e := range examples {
		buf, err := DecodeString(e)
		if err != nil {
			t.Errorf("Decode(%q) failed: %v", e, err)
			continue
		}
		if s := string(buf); s != expected {
			t.Errorf("Decode(%q) = %q, want %q", e, s, expected)
		}
	}
}

func TestNewLineCharacters(t *testing.T) {
	// Each of these should decode to the string "sure", without errors.
	examples := []string{
		"EDTQ4S8",
		"EDTQ4S8\r",
		"EDTQ4S8\n",
		"EDTQ4S8\r\n",
		"EDTQ4S\r\n8",
		"EDT\rQ4S\n8",
		"edt\nq4s\r8",
		"edt\nq4s8",
		"EDTQ4S\n8",
	}
	testStringEncoding(t, "sure", examples)
}

func BenchmarkEncode(b *testing.B) {
	data := make([]byte, 8192)
	buf := make([]byte, EncodedLen(len(data)))
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		Encode(buf, data)
	}
}

func BenchmarkEncodeToString(b *testing.B) {
	data := make([]byte, 8192)
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		EncodeToString(data)
	}
}

func BenchmarkDecode(b *testing.B) {
	data := make([]byte, EncodedLen(8192))
	Encode(data, make([]byte, 8192))
	buf := make([]byte, 8192)
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		Decode(buf, data)
	}
}

func BenchmarkDecodeString(b *testing.B) {
	data := EncodeToString(make([]byte, 8192))
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		DecodeString(data)
	}
}

/* TODO: rewrite without using goroutines
func TestBufferedDecodingSameError(t *testing.T) {
	testcases := []struct {
		prefix            string
		chunkCombinations [][]string
		expected          error
	}{
		// Normal case, this is valid input
		{"helloworld", [][]string{
			{"D1JP", "RV3F", "EXQQ", "4V34"},
			{"D1JPRV3FEXQQ4V34"},
			{"D1J", "PRV", "3FE", "XQQ", "4V3", "4"},
			{"D1JPRV3FEXQQ4V", "34"},
		}, nil},

		// Normal case, this is valid input
		{"fooba", [][]string{
			{"CSQPYRK1"},
			{"CSQPYRK", "1"},
			{"CSQPYR", "K1"},
			{"CSQPY", "RK1"},
			{"CSQPY", "RK", "1"},
			{"CSQPY", "RK1"},
			{"CSQP", "YR", "K1"},
		}, nil},

		// NOTE: many test cases have been removed as we don't return ErrUnexpectedEOF.
	}

	for _, testcase := range testcases {
		for _, chunks := range testcase.chunkCombinations {
			pr, pw := io.Pipe()

			// Write the encoded chunks into the pipe
			go func() {
				for _, chunk := range chunks {
					pw.Write([]byte(chunk))
				}
				pw.Close()
			}()

			decoder := NewDecoder(pr)
			back, err := io.ReadAll(decoder)

			if err != testcase.expected {
				t.Errorf("Expected %v, got %v; case %s %+v", testcase.expected, err, testcase.prefix, chunks)
			}
			if testcase.expected == nil {
				testEqual(t, "Decode from NewDecoder(chunkReader(%v)) = %q, want %q", chunks, string(back), testcase.prefix)
			}
		}
	}
}
*/

func TestEncodedLen(t *testing.T) {
	type test struct {
		n    int
		want int64
	}
	tests := []test{
		{0, 0},
		{1, 2},
		{2, 4},
		{3, 5},
		{4, 7},
		{5, 8},
		{6, 10},
		{7, 12},
		{10, 16},
		{11, 18},
	}
	// check overflow
	tests = append(tests, test{(math.MaxInt-4)/8 + 1, 1844674407370955162})
	tests = append(tests, test{math.MaxInt/8*5 + 4, math.MaxInt})
	for _, tt := range tests {
		if got := EncodedLen(tt.n); int64(got) != tt.want {
			t.Errorf("EncodedLen(%d): got %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestDecodedLen(t *testing.T) {
	type test struct {
		n    int
		want int64
	}
	tests := []test{
		{0, 0},
		{2, 1},
		{4, 2},
		{5, 3},
		{7, 4},
		{8, 5},
		{10, 6},
		{12, 7},
		{16, 10},
		{18, 11},
	}
	// check overflow
	tests = append(tests, test{math.MaxInt/5 + 1, 1152921504606846976})
	tests = append(tests, test{math.MaxInt, 5764607523034234879})
	for _, tt := range tests {
		if got := DecodedLen(tt.n); int64(got) != tt.want {
			t.Errorf("DecodedLen(%d): got %d, want %d", tt.n, got, tt.want)
		}
	}
}
