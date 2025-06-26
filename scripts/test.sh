#!/bin/bash

set -e

echo "Waiting for server to start..."
sleep 5  # Wait for DB and API to boot up

echo "Sending transaction for user 1 (win 10.15)..."
curl -s -X POST http://localhost:8080/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state":"win", "amount":"10.15", "transactionId":"txn_001"}'

echo ""
echo "Fetching balance for user 1..."
curl -s http://localhost:8080/user/1/balance

echo ""
echo "Test completed!"
