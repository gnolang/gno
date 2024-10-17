package blake3

import (
	"encoding/hex"
	"testing"
)

func TestSum256(t *testing.T) {
	got := Sum256([]byte("do not spoil what you have by desiring what you have not"))[:]
	expected := "5bad40bf47e20b26d1b4955464b717545321b1381dc354dcd41a065a8f4f513b"

	if hex.EncodeToString(got) != expected {
		t.Errorf("got %v(%T), expected %v(%T)", hex.EncodeToString(got), got, expected, expected)
	}
}

func TestSum512(t *testing.T) {
	got := Sum512([]byte("do not spoil what you have by desiring what you have not"))[:]
	expected := "5bad40bf47e20b26d1b4955464b717545321b1381dc354dcd41a065a8f4f513be4e747b996ec9100febc5c19ace1eba197aacf2250b466213de14a04d82c5132"

	if hex.EncodeToString(got) != expected {
		t.Errorf("got %v(%T), expected %v(%T)", hex.EncodeToString(got), got, expected, expected)
	}
}
