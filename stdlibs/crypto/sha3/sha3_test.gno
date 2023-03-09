package sha3

import (
	"std"
	"testing"

	"crypto/sha3"
	"encoding/hex"
)

func TestSum256(t *testing.T) {
	got := sha3.Sum256([]byte("welcome to gnoland"))
	expected := "40b3c0de951082601cf10ebab6ec23ee2bd91f978a152f295451b9dcbed33ddb"

	if (hex.EncodeToString(got)) != expected {
		t.Errorf("got %v(%T), expected %v(%T)", hex.EncodeToString(got), hex.EncodeToString(got), expected, expected)
	}
}

func TestSum512(t *testing.T) {
	got := sha3.Sum512([]byte("hello gnomes"))
	expected := "efcdb0b2a91166affe1ae1ac437f21edb9dc319808af9bdb59d52e56f68b83b671a3446412df6b931435f7a3e84713069b18e4d23bd8d230ea5bbfec6f59f417"

	if (hex.EncodeToString(got)) != expected {
		t.Errorf("got %v(%T), expected %v(%T)", hex.EncodeToString(got), hex.EncodeToString(got), expected, expected)
	}
}