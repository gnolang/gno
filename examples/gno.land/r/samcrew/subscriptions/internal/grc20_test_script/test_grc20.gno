package main

import (
	"chain"
	"time"

	"gno.land/r/samcrew/subscriptions"
)

// Create a GRC20-based subscription service
// To execute this function, you must use a MsgRun `gnokey maketx run`
//
// IMPORTANT: Before subscribing with GRC20 tokens, you must:
// 1. First approve the subscription realm to spend your tokens
// 2. Then call Subscribe or Topup
//
// Example workflow:
/*
# Step 1: Create the service (this script)
# Note: The service will be owned by the address of KEY
gnokey maketx run \
  --gas-fee 1000000ugnot \
  --gas-wanted 100000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY \
  ./internal/grc20_test_script/test_grc20.gno

# Step 2 (Testing only): Get some test tokens from the faucet
# This step is only needed if you're testing and don't have tokens yet
gnokey maketx call \
  -pkgpath "gno.land/r/demo/defi/foo20" \
  -func "Faucet" \
  --gas-fee 1000000ugnot \
  --gas-wanted 5000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY

# Step 3: Approve the subscription realm to spend your GRC20 tokens
# First, get the subscription realm's address with:
gnokey query vm/qeval -data 'gno.land/r/samcrew/subscriptions.GetRealmAddress()' --remote localhost:26657
# Then use that address in the Approve command below
gnokey maketx call \
  -pkgpath "gno.land/r/demo/defi/foo20" \
  -func "Approve" \
  -args "<ADDRESS>" \
  -args "50000000" \
  --gas-fee 1000000ugnot \
  --gas-wanted 5000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY

# Step 4: Subscribe to the service with GRC20 tokens
# Service key format: <OWNER_ADDRESS>:<SERVICE_NAME>
# Replace <OWNER_ADDRESS> with the address of the service owner (from Step 1)
gnokey maketx call \
  -pkgpath "gno.land/r/samcrew/subscriptions" \
  -func "Subscribe" \
  -args "<OWNER_ADDRESS>:GRC20 Premium Plan" \
  -args "gno.land/r/demo/defi/foo20" \
  -args "5000000" \
  --gas-fee 1000000ugnot \
  --gas-wanted 5000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY

# Step 5 (Optional): Top up with more GRC20 tokens
# First approve again if needed, then:
gnokey maketx call \
  -pkgpath "gno.land/r/samcrew/subscriptions" \
  -func "Topup" \
  -args "<OWNER_ADDRESS>:GRC20 Premium Plan" \
  -args "gno.land/r/demo/defi/foo20" \
  -args "3000000" \
  --gas-fee 1000000ugnot \
  --gas-wanted 10000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY
*/
func main() {
	// Create a new service with GRC20 token pricing
	// Service costs 5 FOO20 tokens per week
	subscriptions.NewService(
		cross,
		"GRC20 Premium Plan",
		"Weekly subscription paid with GRC20 tokens (FOO20)",
		7*24*time.Hour, // 7 days (1 week)
		chain.Coins{},  // No native coins
		chain.Coins{{Denom: "gno.land/r/demo/defi/foo20", Amount: 5000000}}, // 5 FOO20 tokens
	)
}
