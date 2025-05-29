package ctg

import (
	"std"
	"testing"
)

func TestConvertCosmosToGno(t *testing.T) {
	decoded := [20]byte{
		0x01, 0x02, 0x03, 0x04, 0x05,
		0x06, 0x07, 0x08, 0x09, 0x0a,
		0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14,
	}

	cosmosAddr := std.EncodeBech32("cosmos", decoded)
	wantGno := std.EncodeBech32("g", decoded)

	got, err := ConvertCosmosToGno(cosmosAddr.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != wantGno {
		t.Fatalf("got %s, want %s", got, wantGno)
	}

	// invalid bech32
	if _, err := ConvertCosmosToGno("not-bech32"); err == nil {
		t.Fatalf("expected error for invalid bech32")
	}

	// wrong prefix
	gAddr := std.EncodeBech32("g", decoded)
	if _, err := ConvertCosmosToGno(gAddr.String()); err == nil {
		t.Fatalf("expected error for non-cosmos prefix")
	}
}
