package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	tm2Client "github.com/gnolang/faucet/client/http"
	"github.com/stretchr/testify/assert"
)

func mockedCheckAccountBalance(amount int64, err error) func(tm2Client *tm2Client.Client, walletAddress string) (int64, error) {
	return func(tm2Client *tm2Client.Client, walletAddress string) (int64, error) {
		return amount, err
	}
}

func TestGetAccountBalanceMiddleware(t *testing.T) {
	maxBalance := int64(1000)

	tests := []struct {
		name             string
		requestBody      map[string]string
		expectedStatus   int
		expectedBody     string
		checkBalanceFunc func(tm2Client *tm2Client.Client, walletAddress string) (int64, error)
	}{
		{
			name:             "Valid address with low balance (should pass)",
			requestBody:      map[string]string{"to": "valid_address_low_balance"},
			expectedStatus:   http.StatusOK,
			expectedBody:     "next handler reached",
			checkBalanceFunc: mockedCheckAccountBalance(500, nil),
		},
		{
			name:             "Valid address with high balance (should fail)",
			requestBody:      map[string]string{"To": "valid_address_high_balance"},
			expectedStatus:   http.StatusBadRequest,
			expectedBody:     "accounts is already topped up",
			checkBalanceFunc: mockedCheckAccountBalance(2*maxBalance, nil),
		},
		{
			name:             "Invalid address (should fail)",
			requestBody:      map[string]string{"To": "invalid_address"},
			expectedStatus:   http.StatusBadRequest,
			expectedBody:     "account not found",
			checkBalanceFunc: mockedCheckAccountBalance(2*maxBalance, errors.New("account not found")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkAccountBalance = tt.checkBalanceFunc
			// Convert request body to JSON
			reqBody, _ := json.Marshal(tt.requestBody)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/claim", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create ResponseRecorder
			rr := httptest.NewRecorder()

			// Mock next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("next handler reached"))
			})

			// Apply middleware
			handler := getAccountBalanceMiddleware(nil, maxBalance)(nextHandler)
			handler.ServeHTTP(rr, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
