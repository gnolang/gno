package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tendermint/classic/libs/common"
	tmtypes "github.com/tendermint/classic/types"
	"github.com/tendermint/go-amino-x"

	authtypes "github.com/tendermint/classic/sdk/x/auth/types"
	stakingtypes "github.com/tendermint/classic/sdk/x/staking/types"
)

// GenesisState defines the raw genesis transaction in JSON
type GenesisState struct {
	GenTxs []json.RawMessage `json:"gentxs" yaml:"gentxs"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(genTxs []json.RawMessage) GenesisState {
	return GenesisState{
		GenTxs: genTxs,
	}
}

// NewGenesisStateFromStdTx creates a new GenesisState object
// from auth transactions
func NewGenesisStateFromStdTx(genTxs []authtypes.StdTx) GenesisState {
	genTxsBz := make([]json.RawMessage, len(genTxs))
	for i, genTx := range genTxs {
		genTxsBz[i] = amino.MustMarshalJSON(genTx)
	}
	return NewGenesisState(genTxsBz)
}

// GetGenesisStateFromAppState gets the genutil genesis state from the expected app state
func GetGenesisStateFromAppState(appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		amino.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}
	return genesisState
}

// SetGenesisStateInAppState sets the genutil genesis state within the expected app state
func SetGenesisStateInAppState(
	appState map[string]json.RawMessage, genesisState GenesisState) map[string]json.RawMessage {

	genesisStateBz := amino.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz
	return appState
}

// GenesisStateFromGenDoc creates the core parameters for genesis initialization
// for the application.
//
// NOTE: The pubkey input is this machines pubkey.
func GenesisStateFromGenDoc(genDoc tmtypes.GenesisDoc,
) (genesisState map[string]json.RawMessage, err error) {

	if err = amino.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		return genesisState, err
	}
	return genesisState, nil
}

// GenesisStateFromGenFile creates the core parameters for genesis initialization
// for the application.
//
// NOTE: The pubkey input is this machines pubkey.
func GenesisStateFromGenFile(genFile string,
) (genesisState map[string]json.RawMessage, genDoc *tmtypes.GenesisDoc, err error) {

	if !common.FileExists(genFile) {
		return genesisState, genDoc,
			fmt.Errorf("%s does not exist, run `init` first", genFile)
	}
	genDoc, err = tmtypes.GenesisDocFromFile(genFile)
	if err != nil {
		return genesisState, genDoc, err
	}

	genesisState, err = GenesisStateFromGenDoc(*genDoc)
	return genesisState, genDoc, err
}

// ValidateGenesis validates GenTx transactions
func ValidateGenesis(genesisState GenesisState) error {
	for i, genTx := range genesisState.GenTxs {
		var tx authtypes.StdTx
		if err := amino.UnmarshalJSON(genTx, &tx); err != nil {
			return err
		}

		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return errors.New(
				"must provide genesis StdTx with exactly 1 CreateValidator message")
		}

		// TODO: abstract back to staking
		if _, ok := msgs[0].(stakingtypes.MsgCreateValidator); !ok {
			return fmt.Errorf(
				"genesis transaction %v does not contain a MsgCreateValidator", i)
		}
	}
	return nil
}
