package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

func TestSignED25519Empty(t *testing.T) {
	ledger := &tendermintLedger{api: &mockLedgerDevice{}}
	if _, err := ledger.signED25519(nil); err == nil {
		t.Fatalf("expected error for empty sign bytes")
	}
}

func TestSignED25519Chunking(t *testing.T) {
	message := make([]byte, ledgerMessageChunkSize+8)
	for i := range message {
		message[i] = byte(i)
	}

	signature := make([]byte, ed25519.SignatureSize)
	device := &mockLedgerDevice{}
	device.exchange = func([]byte) ([]byte, error) {
		if len(device.calls) == 2 {
			return signature, nil
		}
		return []byte{0x00}, nil
	}

	ledger := &tendermintLedger{api: device}
	resp, err := ledger.signED25519(message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != ed25519.SignatureSize {
		t.Fatalf("unexpected signature length: %d", len(resp))
	}

	if len(device.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(device.calls))
	}

	first := device.calls[0]
	if len(first) != 5+ledgerMessageChunkSize {
		t.Fatalf("unexpected first call length: %d", len(first))
	}
	if first[0] != ledgerCLA || first[1] != ledgerINSSignED25519 || first[2] != 1 || first[3] != 2 {
		t.Fatalf("unexpected header for first call: %x", first[:5])
	}

	second := device.calls[1]
	if second[2] != 2 || second[3] != 2 {
		t.Fatalf("unexpected header for second call: %x", second[:5])
	}
	if second[4] != 8 {
		t.Fatalf("unexpected chunk length: %d", second[4])
	}
}

func TestSignED25519ExchangeError(t *testing.T) {
	device := &mockLedgerDevice{
		exchange: func([]byte) ([]byte, error) {
			return []byte{0x42}, errors.New("boom")
		},
	}
	ledger := &tendermintLedger{api: device}

	_, err := ledger.signED25519([]byte{0x01})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "0x42") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublicKeyHeaders(t *testing.T) {
	expected := make([]byte, ed25519.PubKeyEd25519Size)
	for i := range expected {
		expected[i] = byte(i)
	}

	device := &mockLedgerDevice{
		exchange: func(command []byte) ([]byte, error) {
			return expected, nil
		},
	}
	ledger := &tendermintLedger{api: device}

	pubKeyBytes, err := ledger.getPublicKeyED25519()
	if err != nil {
		t.Fatalf("unexpected pubkey error: %v", err)
	}

	if len(pubKeyBytes) != ed25519.PubKeyEd25519Size {
		t.Fatalf("unexpected pubkey length: %d", len(pubKeyBytes))
	}
	for i := range expected {
		if pubKeyBytes[i] != expected[i] {
			t.Fatalf("unexpected pubkey byte at %d: %x", i, pubKeyBytes[i])
		}
	}

	if len(device.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(device.calls))
	}

	pubKeyCmd := device.calls[0]
	if len(pubKeyCmd) != 5 || pubKeyCmd[0] != ledgerCLA || pubKeyCmd[1] != ledgerINSPublicKeyED25519 {
		t.Fatalf("unexpected pubkey command: %x", pubKeyCmd)
	}
}
