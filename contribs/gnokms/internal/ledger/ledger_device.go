package ledger

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zondax/hid"
	ledger_go "github.com/zondax/ledger-go"
)

const (
	ledgerCLA           = 0x56
	ledgerINSGetVersion = 0x00
)

type ledgerAppVersion struct {
	AppMode byte
	Major   byte
	Minor   byte
	Patch   byte
}

func requiredLedgerVersion() ledgerAppVersion {
	return ledgerAppVersion{AppMode: 0, Major: 0, Minor: 5, Patch: 0}
}

func (v ledgerAppVersion) meetsMinimum(minimum ledgerAppVersion) bool {
	if v.Major != minimum.Major {
		return v.Major > minimum.Major
	}
	if v.Minor != minimum.Minor {
		return v.Minor > minimum.Minor
	}
	return v.Patch >= minimum.Patch
}

type tendermintLedger struct {
	api ledger_go.LedgerDevice
}

func openTendermintLedger() (*tendermintLedger, error) {
	if !hid.Supported() {
		return nil, errors.New("ledger support is not enabled, try building with CGO_ENABLED=1")
	}

	ledgerAdmin := ledger_go.NewLedgerAdmin()
	ledgerAPI, err := ledgerAdmin.Connect(0)
	if err != nil {
		return nil, err
	}

	return &tendermintLedger{api: ledgerAPI}, nil
}

func (ledger *tendermintLedger) Close() error {
	return ledger.api.Close()
}

func (ledger *tendermintLedger) getVersion() (*ledgerAppVersion, error) {
	message := []byte{ledgerCLA, ledgerINSGetVersion, 0, 0, 0}
	response, err := ledger.api.Exchange(message)
	if err != nil {
		return nil, err
	}

	if len(response) < 4 {
		return nil, errors.New("invalid response")
	}

	return &ledgerAppVersion{
		AppMode: response[0],
		Major:   response[1],
		Minor:   response[2],
		Patch:   response[3],
	}, nil
}

func validateLedgerApp(ledger *tendermintLedger) error {
	version, err := ledger.getVersion()
	if err != nil {
		if strings.Contains(err.Error(), "APDU_CODE_CLA_NOT_SUPPORTED") {
			return errors.New("are you sure the Tendermint Validator app is open?")
		}
		return err
	}

	if !version.meetsMinimum(requiredLedgerVersion()) {
		req := requiredLedgerVersion()
		return fmt.Errorf(
			"ledger app version %d.%d.%d is below required %d.%d.%d",
			version.Major, version.Minor, version.Patch,
			req.Major, req.Minor, req.Patch,
		)
	}

	return nil
}
