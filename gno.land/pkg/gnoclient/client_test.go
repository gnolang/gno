package gnoclient

import (
	"testing"
)

func TestClient_Request(t *testing.T) {
	client := Client{
		Remote:  "localhost:12345",
		ChainID: "test",
	}
	_ = client

	// TODO: xxx
}
