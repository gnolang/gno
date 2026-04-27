#!/bin/sh

echo "Starting e2e tests..."

# Define test node URL and key info
NODE_URL="http://gnoland:26657"
TEST1_ADDR="g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
TEST1_MNEMONIC="source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
# TODO: make the test even more e2e by generating random account and using a faucet.

echo "Waiting for gnoland to be ready..."

# Retry mechanism to wait for node to be ready
MAX_RETRIES=40
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    echo "Attempting to connect to $NODE_URL... (attempt $((RETRY_COUNT + 1))/$MAX_RETRIES)"
    
    # Try to query and capture the output
    OUTPUT=$(/usr/bin/gnokey query bank/balances/"$TEST1_ADDR" -remote="$NODE_URL" 2>&1)
    EXIT_CODE=$?
    
    if [ $EXIT_CODE -eq 0 ]; then
        echo "Successfully connected! Response:"
        echo "$OUTPUT"
        echo "Node is ready!"
        break
    else
        echo "Failed with exit code $EXIT_CODE. Output:"
        echo "$OUTPUT"
    fi
    
    echo "Waiting 3 seconds before retry..."
    sleep 3
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "Failed to connect to node after $MAX_RETRIES attempts"
    exit 1
fi

echo "=== Running Query Tests ==="

# Query node status using ABCI info
echo "1. Checking if node is ready..."
# Try a simple query to check if node is responding
/usr/bin/gnokey query bank/balances/"$TEST1_ADDR" -remote="$NODE_URL" || { echo "Failed to connect to node"; exit 1; }

# Query account information
echo "2. Querying test1 account information..."
/usr/bin/gnokey query auth/accounts/"$TEST1_ADDR" -remote="$NODE_URL" || { echo "Failed to query account"; exit 1; }

echo "=== Running Transaction Tests ==="

# Import test1 key
echo "3. Importing test1 key..."
# First, remove the key if it exists (ignore errors)
/usr/bin/gnokey delete test1 -home /tmp/gnokey -force 2>/dev/null || true
# Use printf to provide password twice (for password and confirmation)
printf "%s\ntest1234\ntest1234\n" "$TEST1_MNEMONIC" | /usr/bin/gnokey add test1 -recover -insecure-password-stdin=true -home /tmp/gnokey || { echo "Failed to import test1 key"; exit 1; }

# Make a simple transaction - send some coins
echo "4. Sending coins to another address..."
echo "test1234" | /usr/bin/gnokey maketx send \
  -to "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" \
  -send "100ugnot" \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  -broadcast \
  -chainid test \
  -remote "$NODE_URL" \
  -insecure-password-stdin=true \
  -home /tmp/gnokey \
  test1 || { echo "Failed to send coins"; exit 1; }

echo "=== Running Additional Transaction Tests ==="

# Check balance after first transaction
echo "5. Checking test1 balance after sending coins..."
/usr/bin/gnokey query bank/balances/"$TEST1_ADDR" -remote="$NODE_URL" || { echo "Failed to query balance"; exit 1; }

# Use a predefined test2 address
echo "6. Using predefined test2 address..."
TEST2_ADDR="g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
echo "Test2 address: $TEST2_ADDR"

# Send coins from test1 to test2
echo "7. Sending coins from test1 to test2..."
echo "test1234" | /usr/bin/gnokey maketx send \
  -to "$TEST2_ADDR" \
  -send "500ugnot" \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  -broadcast \
  -chainid test \
  -remote "$NODE_URL" \
  -insecure-password-stdin=true \
  -home /tmp/gnokey \
  test1 || { echo "Failed to send coins to test2"; exit 1; }

# Check test2 balance
echo "8. Checking test2 balance..."
/usr/bin/gnokey query bank/balances/"$TEST2_ADDR" -remote="$NODE_URL" || { echo "Failed to query test2 balance"; exit 1; }

# Check test1 final balance
echo "9. Checking test1 final balance..."
/usr/bin/gnokey query bank/balances/"$TEST1_ADDR" -remote="$NODE_URL" || { echo "Failed to query test1 final balance"; exit 1; }

# TODO:
# - interact with existing contracts
# - upload a new contract and interact with it

echo "All tests passed successfully!"
