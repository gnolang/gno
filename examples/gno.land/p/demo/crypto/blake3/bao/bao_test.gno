package bao_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"gno.land/p/demo/crypto/blake3"
	"gno.land/p/demo/crypto/blake3/bao"
)

func toHex(data []byte) string { return hex.EncodeToString(data) }

func genTestData() []byte {
	data := make([]byte, 1<<14)
	blake3.New(0, nil).XOF().Read(data)
	return data
}

// XXX: can't read from files in Gno. Commenting out the test.
// func TestBaoGolden(t *testing.T) {
// 	data, err := os.ReadFile("../testdata/vectors.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	goldenInterleaved, err := os.ReadFile("../testdata/bao-golden.bao")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	goldenOutboard, err := os.ReadFile("../testdata/bao-golden.obao")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	interleaved, root := bao.EncodeBuf(data, 0, false)
// 	if toHex(root[:]) != "6654fbd1836b531b25e2782c9cc9b792c80abb36b024f59db5d5f6bd3187ddfe" {
// 		t.Errorf("bad root: %x", root)
// 	} else if !bytes.Equal(interleaved, goldenInterleaved) {
// 		t.Error("bad interleaved encoding")
// 	}

// 	outboard, root := bao.EncodeBuf(data, 0, true)
// 	if toHex(root[:]) != "6654fbd1836b531b25e2782c9cc9b792c80abb36b024f59db5d5f6bd3187ddfe" {
// 		t.Errorf("bad root: %x", root)
// 	} else if !bytes.Equal(outboard, goldenOutboard) {
// 		t.Error("bad outboard encoding")
// 	}

// 	// test empty input
// 	interleaved, root = bao.EncodeBuf(nil, 0, false)
// 	if toHex(root[:]) != "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262" {
// 		t.Errorf("bad root: %x", root)
// 	} else if toHex(interleaved[:]) != "0000000000000000" {
// 		t.Errorf("bad interleaved encoding: %x", interleaved)
// 	} else if !bao.VerifyBuf(interleaved, nil, 0, root) {
// 		t.Error("verify failed")
// 	}
// 	outboard, root = bao.EncodeBuf(nil, 0, true)
// 	if toHex(root[:]) != "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262" {
// 		t.Errorf("bad root: %x", root)
// 	} else if toHex(outboard[:]) != "0000000000000000" {
// 		t.Errorf("bad outboard encoding: %x", outboard)
// 	} else if !bao.VerifyBuf(nil, outboard, 0, root) {
// 		t.Error("verify failed")
// 	}
// }

func TestBaoInterleaved(t *testing.T) {
	data := genTestData()

	for group := 0; group < 3; group++ {
		interleaved, root := bao.EncodeBuf(data, group, false)
		if !bao.VerifyBuf(interleaved, nil, group, root) {
			t.Fatal("verify failed")
		}
		badRoot := root
		badRoot[0] ^= 1
		if bao.VerifyBuf(interleaved, nil, group, badRoot) {
			t.Fatal("verify succeeded with bad root")
		}
		badPrefix := append([]byte(nil), interleaved...)
		badPrefix[0] ^= 1
		if bao.VerifyBuf(badPrefix, nil, group, root) {
			t.Fatal("verify succeeded with bad length prefix")
		}
		badCVs := append([]byte(nil), interleaved...)
		badCVs[8] ^= 1
		if bao.VerifyBuf(badCVs, nil, group, root) {
			t.Fatal("verify succeeded with bad cv data")
		}
		badData := append([]byte(nil), interleaved...)
		badData[len(badData)-1] ^= 1
		if bao.VerifyBuf(badData, nil, group, root) {
			t.Fatal("verify succeeded with bad content")
		}
		extraData := append(append([]byte(nil), interleaved...), 1, 2, 3)
		if bao.VerifyBuf(extraData, nil, group, root) {
			t.Fatal("verify succeeded with extra data")
		}
	}
}

func TestBaoOutboard(t *testing.T) {
	data := genTestData()

	for group := 0; group < 3; group++ {
		outboard, root := bao.EncodeBuf(data, group, true)
		if !bao.VerifyBuf(data, outboard, group, root) {
			t.Fatal("verify failed")
		}
		badRoot := root
		badRoot[0] ^= 1
		if bao.VerifyBuf(data, outboard, group, badRoot) {
			t.Fatal("verify succeeded with bad root")
		}
		badPrefix := append([]byte(nil), outboard...)
		badPrefix[0] ^= 1
		if bao.VerifyBuf(data, badPrefix, group, root) {
			t.Fatal("verify succeeded with bad length prefix")
		}
		badCVs := append([]byte(nil), outboard...)
		badCVs[8] ^= 1
		if bao.VerifyBuf(data, badCVs, group, root) {
			t.Fatal("verify succeeded with bad cv data")
		}
	}
}

func TestBaoChunkGroup(t *testing.T) {
	// from https://github.com/n0-computer/abao/blob/9b756ec8097afc782d76f7aec0a5ac9f4b82329a/tests/test_vectors.json
	const group = 4 // 16 KiB
	baoInput := func(n int) (in []byte) {
		for i := uint32(1); len(in) < n; i++ {
			in = binary.LittleEndian.AppendUint32(in, i)
		}
		return in[:n]
	}
	for _, test := range []struct {
		inputLen int
		exp      string
	}{
		{0, "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"},
		{1, "48fc721fbbc172e0925fa27af1671de225ba927134802998b10a1568a188652b"},
		{1023, "15f8c1ae1049fe7e837186612c8ce732e66835841a4569b71e4ac3e3d3411b90"},
		{1024, "f749c19181983b839cd97fe121cebaf076bc951e8c8e6d64accfedad5951ec22"},
		{1025, "3613596275c4ea790774dedf20835b2daf86cacc892feef6ce720c121572f1f9"},
		{16383, "f0970fbfe2f1c5145fa6aa31833779803d5c53743a8443ed1395218f511834ba"},
		{16384, "b318758645c4467406c829a5f3da7cab00010fccccf4b7c314525cd85e2d0af8"},
		{16385, "12a6a6b0554e7f3eed485f668bfd3b37382a2beee5e7ed5594c4a91c4c70f4aa"},
		{32768, "8008de557073cab60f851191359ad9dc1afe9dc6152668ee01825c56ac5a754e"},
		{49152, "91823357fefc308b57bb85ebed1d1edeba3c355e804dc63fa98fcb82554b1566"},
		{180224, "4742cbae9485ce1b86ab359c1a84e203f819795d018b22a5c70c5c4577dd732e"},
		{212992, "760c549edfe95c734b1d6a9b846d81692ed3ca022b541442949a0e42fe570df2"},
	} {
		input := baoInput(test.inputLen)
		_, root := bao.EncodeBuf(input, group, false)
		if out := toHex(root[:]); out != test.exp {
			t.Errorf("output %v did not match test vector:\n\texpected: %v...\n\t     got: %v...", test.inputLen, test.exp[:10], out[:10])
		}
	}
}

func TestBaoStreaming(t *testing.T) {
	data := genTestData()

	enc, root := bao.EncodeBuf(data, 0, false)
	if root != blake3.Sum256(data) {
		t.Fatal("bad root")
	}
	var buf bytes.Buffer
	if ok, err := bao.Decode(&buf, bytes.NewReader(enc), nil, 0, root); err != nil || !ok {
		t.Fatal("decode failed")
	} else if !bytes.Equal(buf.Bytes(), data) {
		t.Fatal("bad decode")
	}

	// corrupt root; nothing should be written to buf
	buf.Reset()
	if ok, err := bao.Decode(&buf, bytes.NewReader(enc), nil, 0, [32]byte{}); err != nil {
		t.Fatal("decode failed")
	} else if ok {
		t.Fatal("decode succeeded with bad root")
	} else if buf.Len() != 0 {
		t.Fatal("buf was written with bad root")
	}

	// corrupt a byte halfway through; buf should only be partially written
	buf.Reset()
	enc[len(enc)/2] ^= 1
	if ok, err := bao.Decode(&buf, bytes.NewReader(enc), nil, 0, root); err != nil {
		t.Fatal("decode failed")
	} else if ok {
		t.Fatal("decode succeeded with bad data")
	} else if !bytes.Equal(buf.Bytes(), data[:buf.Len()]) {
		t.Fatal("invalid data was written to buf")
	}
}

func TestBaoSlice(t *testing.T) {
	data := genTestData()

	for _, test := range []struct {
		off, len uint64
	}{
		{0, uint64(len(data))},
		{0, 1024},
		{1024, 1024},
		{0, 10},
		{1020, 10},
		{1030, uint64(len(data) - 1030)},
	} {
		// combined encoding
		{
			enc, root := bao.EncodeBuf(data, 0, false)
			var buf bytes.Buffer
			if err := bao.ExtractSlice(&buf, bytes.NewReader(enc), nil, 0, test.off, test.len); err != nil {
				t.Error(err)
			} else if vdata, ok := bao.VerifySlice(buf.Bytes(), 0, test.off, test.len, root); !ok {
				t.Error("combined verify failed", test)
			} else if !bytes.Equal(vdata, data[test.off:][:test.len]) {
				t.Error("combined bad decode", test, vdata, data[test.off:][:test.len])
			}
		}
		// outboard encoding
		{
			enc, root := bao.EncodeBuf(data, 0, true)
			start, end := (test.off/1024)*1024, ((test.off+test.len+1024-1)/1024)*1024
			if end > uint64(len(data)) {
				end = uint64(len(data))
			}
			var buf bytes.Buffer
			if err := bao.ExtractSlice(&buf, bytes.NewReader(data[start:end]), bytes.NewReader(enc), 0, test.off, test.len); err != nil {
				t.Error(err)
			} else if vdata, ok := bao.VerifySlice(buf.Bytes(), 0, test.off, test.len, root); !ok {
				t.Error("outboard verify failed", test)
			} else if !bytes.Equal(vdata, data[test.off:][:test.len]) {
				t.Error("outboard bad decode", test, vdata, data[test.off:][:test.len])
			}
		}
	}
}
