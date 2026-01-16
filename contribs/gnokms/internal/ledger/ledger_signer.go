// Inspired by https://github.com/cosmos/ledger-cosmos-go/blob/9e3c918a05e0f84ce4205215d71c731f23888ec9/validator_app.go

package ledger

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

const (
	ledgerINSPublicKeyED25519 = 0x01
	ledgerINSSignED25519      = 0x02
	ledgerMessageChunkSize    = 32 // needded for macOS HID IO issues
)

func (ledger *tendermintLedger) signED25519(message []byte) ([]byte, error) {
	if len(message) == 0 {
		return nil, errors.New("sign bytes are empty")
	}

	var packetIndex byte = 1
	packetCount := byte(math.Ceil(float64(len(message)) / float64(ledgerMessageChunkSize)))

	var finalResponse []byte

	for packetIndex <= packetCount {
		chunk := ledgerMessageChunkSize
		if len(message) < ledgerMessageChunkSize {
			chunk = len(message)
		}

		header := []byte{
			ledgerCLA,
			ledgerINSSignED25519,
			packetIndex,
			packetCount,
			byte(chunk),
		}

		apduMessage := append(header, message[:chunk]...)

		response, err := ledger.api.Exchange(apduMessage)
		if err != nil {
			return nil, formatLedgerError(err, response)
		}

		finalResponse = response
		message = message[chunk:]
		packetIndex++
	}

	return finalResponse, nil
}

func (ledger *tendermintLedger) getPublicKeyED25519() ([]byte, error) {
	message := []byte{ledgerCLA, ledgerINSPublicKeyED25519, 0, 0, 0}
	return ledger.api.Exchange(message)
}

func formatLedgerError(err error, response []byte) error {
	if len(response) == 1 {
		return fmt.Errorf("ledger rejected sign bytes (parse error code: 0x%02x): %w", response[0], err)
	}

	return err
}

// ledgerSigner is a gnokms signer backed by the Ledger Tendermint validator app.
type ledgerSigner struct {
	ledger *tendermintLedger
	pubKey ed25519.PubKeyEd25519
	mu     sync.Mutex
}

// ledgerSigner type implements types.Signer.
var _ types.Signer = (*ledgerSigner)(nil)

// PubKey implements types.Signer.
func (ls *ledgerSigner) PubKey() crypto.PubKey {
	return ls.pubKey
}

// Sign implements types.Signer.
func (ls *ledgerSigner) Sign(signBytes []byte) ([]byte, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	signature, err := ls.ledger.signED25519(signBytes)
	if err != nil {
		return nil, err
	}

	if len(signature) != ed25519.SignatureSize {
		return nil, fmt.Errorf("unexpected ledger signature length: %d", len(signature))
	}

	return signature, nil
}

// Close implements types.Signer.
func (ls *ledgerSigner) Close() error {
	return ls.ledger.Close()
}

// newLedgerSigner initializes a new ledger signer using the Tendermint validator app.
func newLedgerSigner() (*ledgerSigner, error) {
	ledger, err := openTendermintLedger()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Ledger Tendermint validator app: %w", err)
	}

	if err := validateLedgerApp(ledger); err != nil {
		_ = ledger.Close()
		return nil, err
	}

	pubKeyBytes, err := ledger.getPublicKeyED25519()
	if err != nil {
		_ = ledger.Close()
		return nil, fmt.Errorf("unable to fetch ledger public key: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PubKeyEd25519Size {
		_ = ledger.Close()
		return nil, fmt.Errorf("unexpected ledger public key length: %d", len(pubKeyBytes))
	}

	var pubKey ed25519.PubKeyEd25519
	copy(pubKey[:], pubKeyBytes)

	return &ledgerSigner{
		ledger: ledger,
		pubKey: pubKey,
	}, nil
}
