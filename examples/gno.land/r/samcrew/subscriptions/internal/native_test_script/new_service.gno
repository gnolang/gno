package main

import (
	"chain"
	"time"

	"gno.land/r/samcrew/subscriptions"
)

// Create a new subscription service with NATIVE coins (ugnot)
// To execute this function, you must use a MsgRun `gnokey maketx run`
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
// Example:
/*
# Step 1: Create the service (this script)
gnokey maketx run \
  --gas-fee 1000000ugnot \
  --gas-wanted 100000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY \
  ./internal/new_service_script/new_service.gno

# Step 2: Subscribe to the service with native coins
gnokey maketx call \
  -pkgpath "gno.land/r/samcrew/subscriptions" \
  -func "Subscribe" \
  -args "Premium Plan" \
  -args "" \
  -args "0" \
  -send "1000000ugnot" \
  --gas-fee 1000000ugnot \
  --gas-wanted 5000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY

# Step 3 (Optional): Top up with more native coins
gnokey maketx call \
  -pkgpath "gno.land/r/samcrew/subscriptions" \
  -func "Topup" \
  -args "Premium Plan" \
  -args "" \
  -args "0" \
  -send "2000000ugnot" \
  --gas-fee 1000000ugnot \
  --gas-wanted 5000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY
*/
func main() {
	// Create a new service with 30-day renewal period
	// Price: 1 GNOT (1000000 ugnot) per month in NATIVE coins
	subscriptions.NewService(
		cross,
		"Premium Plan",
		"Access to premium features with monthly billing (paid in native GNOT)",
		30*24*time.Hour,                                // 30 days
		chain.Coins{{Denom: "ugnot", Amount: 1000000}}, // 1 GNOT in native coins
		chain.Coins{},                                  // No GRC20 tokens
	)
}
