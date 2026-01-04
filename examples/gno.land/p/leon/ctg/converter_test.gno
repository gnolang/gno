package ctg

import (
	"testing"
)

func TestConvertKnownAddress(t *testing.T) {
	const (
		cosmosAddr = "cosmos1jg8mtutu9khhfwc4nxmuhcpftf0pajdh6svrgs"
		gnoAddr    = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	)
	got, err := ConvertCosmosToGno(cosmosAddr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != gnoAddr {
		t.Fatalf("got %s, want %s", got, gnoAddr)
	}
}

func TestConvertCosmosToGno(t *testing.T) {
	decoded := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x00,
	}

	cosmosAddr := mustEncode("cosmos", decoded)
	wantGno := mustEncode("g", decoded)

	got, err := ConvertCosmosToGno(cosmosAddr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != wantGno {
		t.Fatalf("got %s, want %s", got, wantGno)
	}

	// invalid bech32
	if _, err := ConvertCosmosToGno("not-bech32"); err == nil {
		t.Fatalf("expected error for invalid bech32")
	}

	// wrong prefix
	gAddr := mustEncode("g", decoded)
	if _, err := ConvertCosmosToGno(gAddr); err == nil {
		t.Fatalf("expected error for non-cosmos prefix")
	}
}

func TestConvertAnyToGno(t *testing.T) {
	payload := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x00,
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "cosmos→g",
			input: mustEncode("cosmos", payload),
			want:  mustEncode("g", payload),
		},
		{
			name:  "osmo→g",
			input: mustEncode("osmo", payload),
			want:  mustEncode("g", payload),
		},
		{
			name:    "invalid bech32",
			input:   "xyz123",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertAnyToGno(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestConvertGnoToAny(t *testing.T) {
	payload := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x00,
	}

	gno := address(mustEncode("g", payload))

	t.Run("g→cosmos", func(t *testing.T) {
		got, err := ConvertGnoToAny("cosmos", gno)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != mustEncode("cosmos", payload) {
			t.Fatalf("conversion incorrect: %s", got)
		}
	})

	t.Run("g→foobar", func(t *testing.T) {
		got, err := ConvertGnoToAny("foobar", gno)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != mustEncode("foobar", payload) {
			t.Fatalf("conversion incorrect: %s", got)
		}
	})

	t.Run("g→osmo", func(t *testing.T) {
		got, err := ConvertGnoToAny("osmo", gno)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != mustEncode("osmo", payload) {
			t.Fatalf("conversion incorrect: %s", got)
		}
	})

	t.Run("wrong source prefix", func(t *testing.T) {
		cosmos := mustEncode("cosmos", payload)
		if _, err := ConvertGnoToAny("g", address(cosmos)); err == nil {
			t.Fatalf("expected error for non-g source prefix")
		}
	})

	t.Run("invalid bech32", func(t *testing.T) {
		if _, err := ConvertGnoToAny("cosmos", address("nope")); err == nil {
			t.Fatalf("expected error for invalid bech32")
		}
	})
}
