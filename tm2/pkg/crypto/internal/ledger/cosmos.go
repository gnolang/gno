// Copyright (c) Zondax AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file is a client for the Ledger Cosmos app, ported from
// github.com/cosmos/ledger-cosmos-go v1.0.0 (user_app.go, common.go) and the
// chunking of github.com/zondax/ledger-go, onto this package's transport.

package ledger

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// APDU protocol of the Ledger Cosmos app.
const (
	cosmosCLA = 0x55

	cosmosInsGetVersion = 0x00
	cosmosInsSign       = 0x02
	cosmosInsGetAddress = 0x04

	cosmosP1SignInit = 0x00
	cosmosP1SignAdd  = 0x01
	cosmosP1SignLast = 0x02

	cosmosSignModeTextual = 1 // the highest sign mode the app accepts
	cosmosChunkSize       = 48
	cosmosPubKeyLen       = 33
)

// cosmosMinVersion is the oldest Cosmos app version supported, as in
// ledger-cosmos-go: 2.1.0, written as major<<16 | minor<<8 | patch.
const cosmosMinVersion = 0x020100

var errNotCosmosApp = errors.New("the open Ledger app is not the Cosmos app")

// CosmosApp talks to the Ledger Cosmos app. It implements [SECP256K1].
type CosmosApp struct {
	device device
}

// openCosmosApp returns a [CosmosApp] on dev if the Cosmos app is the one
// open, and errNotCosmosApp if the device answered as another app, or the
// dashboard, would.
func openCosmosApp(dev device) (*CosmosApp, error) {
	app := &CosmosApp{device: dev}
	resp, err := app.exchange(cosmosInsGetVersion, 0, 0, nil)
	var status *statusError
	switch {
	case errors.As(err, &status) && status.sw == swDeviceLocked:
		return nil, cosmosError(err)
	case errors.As(err, &status):
		return nil, errNotCosmosApp
	case err != nil:
		return nil, err
	case len(resp) < 4:
		return nil, fmt.Errorf("malformed Cosmos app version of %d bytes", len(resp))
	}
	// app mode | major | minor | patch
	if int(resp[1])<<16|int(resp[2])<<8|int(resp[3]) < cosmosMinVersion {
		return nil, fmt.Errorf("the Ledger Cosmos app %d.%d.%d is too old: 2.1.0 or later is required", resp[1], resp[2], resp[3])
	}
	return app, nil
}

func (a *CosmosApp) exchange(ins, p1, p2 byte, data []byte) ([]byte, error) {
	return a.device.Exchange(append([]byte{cosmosCLA, ins, p1, p2, byte(len(data))}, data...))
}

// cosmosPath encodes a BIP44 path given as unhardened components, the form
// [hd.BIP44Params.DerivationPath] returns: five little-endian words, the
// first three hardened.
func cosmosPath(path []uint32) ([]byte, error) {
	if len(path) != 5 {
		return nil, fmt.Errorf("path must have 5 components, got %d", len(path))
	}
	out := make([]byte, 0, 4*len(path))
	for i, c := range path {
		if c >= 0x8000_0000 {
			return nil, fmt.Errorf("path component %d is out of range", i)
		}
		if i < 3 {
			c |= 0x8000_0000
		}
		out = binary.LittleEndian.AppendUint32(out, c)
	}
	return out, nil
}

// cosmosError makes a status word from the Cosmos app readable. With some
// of them, the app explains the problem in the response data.
func cosmosError(err error) error {
	var status *statusError
	if !errors.As(err, &status) {
		return err
	}
	switch status.sw {
	case swDeviceLocked:
		return fmt.Errorf("the Ledger device is locked: unlock it (0x%04X)", status.sw)
	case 0x6985, 0x6986:
		return fmt.Errorf("rejected on the Ledger device (0x%04X)", status.sw)
	}
	if len(status.data) > 0 {
		return fmt.Errorf("the Ledger Cosmos app reports: %s (0x%04X)", status.data, status.sw)
	}
	return fmt.Errorf("the Ledger Cosmos app answered status word 0x%04X", status.sw)
}

func (a *CosmosApp) getAddress(path []uint32, hrp string, display bool) ([]byte, string, error) {
	if len(hrp) == 0 || len(hrp) > 83 {
		return nil, "", fmt.Errorf("invalid address prefix %q", hrp)
	}
	for _, c := range []byte(hrp) {
		// BIP-173: printable ASCII, lowercase.
		if c < 33 || c > 126 || (c >= 'A' && c <= 'Z') {
			return nil, "", fmt.Errorf("invalid address prefix %q", hrp)
		}
	}
	pathBytes, err := cosmosPath(path)
	if err != nil {
		return nil, "", err
	}
	p1 := byte(0)
	if display {
		p1 = 1
	}
	data := append([]byte{byte(len(hrp))}, hrp...)
	resp, err := a.exchange(cosmosInsGetAddress, p1, 0, append(data, pathBytes...))
	if err != nil {
		return nil, "", cosmosError(err)
	}
	// pubkey (33) | bech32 address: prefix, "1", data, 6-character checksum
	if len(resp) < cosmosPubKeyLen+len(hrp)+1+1+6 {
		return nil, "", fmt.Errorf("malformed Cosmos app address response of %d bytes", len(resp))
	}
	return resp[:cosmosPubKeyLen], string(resp[cosmosPubKeyLen:]), nil
}

// GetPublicKeySECP256K1 returns the compressed public key for path.
func (a *CosmosApp) GetPublicKeySECP256K1(path []uint32) ([]byte, error) {
	pub, _, err := a.getAddress(path, "cosmos", false)
	return pub, err
}

// GetAddressPubKeySECP256K1 shows the address for path, with prefix hrp, on
// the device and returns it with the compressed public key once the user
// approves it.
func (a *CosmosApp) GetAddressPubKeySECP256K1(path []uint32, hrp string) ([]byte, string, error) {
	return a.getAddress(path, hrp, true)
}

// SignSECP256K1 has the user review msg and returns its DER-encoded
// signature. signMode is the app's sign mode: 0 for amino JSON, 1 for
// textual.
func (a *CosmosApp) SignSECP256K1(path []uint32, msg []byte, signMode byte) ([]byte, error) {
	if signMode > cosmosSignModeTextual {
		return nil, fmt.Errorf("the Ledger Cosmos app has no signing mode %d", signMode)
	}
	pathBytes, err := cosmosPath(path)
	if err != nil {
		return nil, err
	}
	chunks := [][]byte{pathBytes}
	for len(msg) > 0 {
		n := min(len(msg), cosmosChunkSize)
		chunks = append(chunks, msg[:n])
		msg = msg[n:]
	}
	var resp []byte
	for i, chunk := range chunks {
		p1 := byte(cosmosP1SignAdd)
		switch {
		case i == 0:
			p1 = cosmosP1SignInit
		case i == len(chunks)-1:
			p1 = cosmosP1SignLast
		}
		if resp, err = a.exchange(cosmosInsSign, p1, signMode, chunk); err != nil {
			return nil, cosmosError(err)
		}
	}
	return resp, nil
}

// Close closes the connection to the device.
func (a *CosmosApp) Close() error {
	return a.device.Close()
}
