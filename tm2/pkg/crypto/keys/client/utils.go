package client

import (
	"encoding/json"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

const (
	TEXT_FORMAT = "text"
	JSON_FORMAT = "json"
)

func formatQueryResponse(res abci.ResponseQuery) string {
	data := json.RawMessage(res.Data)

	// Create a struct to hold the final JSON structure with ordered fields
	formattedData := struct {
		Height int64           `json:"height"`
		Data   json.RawMessage `json:"data"`
	}{
		Height: res.Height,
		Data:   data,
	}

	// Marshal the final struct into an indented JSON string for readability
	formattedResponse, err := json.MarshalIndent(formattedData, "", "  ")
	if err != nil {
		return fmt.Sprintf("height: %d\ndata: %s\n", res.Height, string(res.Data))
	}

	// Return the formatted JSON string
	return string(formattedResponse)
}

func formatDeliverTxResponse(res abci.ResponseDeliverTx, height int64) string {
	data := json.RawMessage(res.Data)
	events := json.RawMessage(res.EncodeEvents())

	// Create a struct to hold the final JSON structure with ordered fields
	formattedData := struct {
		Data      json.RawMessage `json:"DATA"`
		Status    string          `json:"STATUS"`
		GasWanted int64           `json:"GAS WANTED"`
		GasUsed   int64           `json:"GAS USED"`
		Height    int64           `json:"HEIGHT"`
		Events    json.RawMessage `json:"EVENTS"`
	}{
		Data:      data,
		Status:    "OK!",
		GasWanted: res.GasWanted,
		GasUsed:   res.GasUsed,
		Height:    height,
		Events:    events,
	}

	// Marshal the final struct into an indented JSON string for readability
	formattedResponse, err := json.MarshalIndent(formattedData, "", "  ")
	if err != nil {
		return fmt.Sprintf("Data: %s\nOK!\nGAS WANTED: %d\nGAS USED: %d\nHEIGHT: %d\nEVENTS: %s\n",
			string(res.Data),
			res.GasWanted,
			res.GasUsed,
			height,
			string(res.EncodeEvents()))
	}

	// Return the formatted JSON string
	return string(formattedResponse)
}
