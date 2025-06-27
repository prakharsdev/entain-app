#!/bin/bash

set -e

echo "Waiting for server to start on port 8080..."
for i in {1..20}; do
  if curl -s http://localhost:8080/user/1/balance > /dev/null; then
    echo "Server is up!"
    break
  fi
  echo "Still waiting for server... retry $i"
  sleep 1
done

# Generate a unique transaction ID using current timestamp
TXN_ID="txn_$(date +%s%N)"
echo "Sending transaction for user 1 (win 10.15) with ID: $TXN_ID..."

RESPONSE=$(curl -s -w '\n%{http_code}' -X POST http://localhost:8080/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d "{\"state\":\"win\", \"amount\":\"10.15\", \"transactionId\":\"$TXN_ID\"}")

# Separate body and status
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n1)

echo "Response ($STATUS): $BODY"

echo ""
echo "Fetching balance for user 1..."
curl -s http://localhost:8080/user/1/balance

echo ""
echo "Test completed!"
