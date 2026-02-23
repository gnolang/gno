package ledger

import (
	"errors"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	var gotCommand []byte
	device := &mockLedgerDevice{
		exchange: func(command []byte) ([]byte, error) {
			gotCommand = append([]byte(nil), command...)
			return []byte{0x00, 0x01, 0x02, 0x03}, nil
		},
	}
	ledger := &tendermintLedger{api: device}

	version, err := ledger.getVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gotCommand) != 5 {
		t.Fatalf("unexpected command length: %d", len(gotCommand))
	}
	if gotCommand[0] != ledgerCLA || gotCommand[1] != ledgerINSGetVersion || gotCommand[2] != 0 || gotCommand[3] != 0 || gotCommand[4] != 0 {
		t.Fatalf("unexpected command: %x", gotCommand)
	}

	if version.AppMode != 0x00 || version.Major != 0x01 || version.Minor != 0x02 || version.Patch != 0x03 {
		t.Fatalf("unexpected version: %+v", version)
	}
}

func TestGetVersionInvalidResponse(t *testing.T) {
	device := &mockLedgerDevice{
		exchange: func([]byte) ([]byte, error) {
			return []byte{0x00, 0x01, 0x02}, nil
		},
	}
	ledger := &tendermintLedger{api: device}

	if _, err := ledger.getVersion(); err == nil {
		t.Fatalf("expected error for short response")
	}
}

func TestValidateLedgerAppCLANotSupported(t *testing.T) {
	device := &mockLedgerDevice{
		exchange: func([]byte) ([]byte, error) {
			return nil, errors.New("[APDU_CODE_CLA_NOT_SUPPORTED] Class not supported")
		},
	}
	ledger := &tendermintLedger{api: device}

	err := validateLedgerApp(ledger)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "Tendermint Validator app is open") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateLedgerAppVersionTooLow(t *testing.T) {
	device := &mockLedgerDevice{
		exchange: func([]byte) ([]byte, error) {
			return []byte{0x00, 0x00, 0x04, 0x00}, nil
		},
	}
	ledger := &tendermintLedger{api: device}

	err := validateLedgerApp(ledger)
	if err == nil {
		t.Fatalf("expected version error")
	}
	if !strings.Contains(err.Error(), "below required") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
