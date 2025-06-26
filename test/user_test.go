package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestTransactionAndBalance(t *testing.T) {
	baseURL := "http://localhost:8080"
	userID := 1

	// Step 1: Get initial balance
	resp, err := http.Get(fmt.Sprintf("%s/user/%d/balance", baseURL, userID))
	if err != nil {
		t.Fatalf("Failed to fetch initial balance: %v", err)
	}
	defer resp.Body.Close()

	var initial map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&initial)

	initialBalance, ok := initial["balance"].(string)
	if !ok {
		t.Fatalf("Initial balance not found or invalid")
	}

	var before float64
	fmt.Sscanf(initialBalance, "%f", &before)

	// Step 2: Post a unique transaction
	txnID := fmt.Sprintf("test_txn_%d", time.Now().UnixNano())
	payload := map[string]string{
		"state":         "win",
		"amount":        "5.00",
		"transactionId": txnID,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/user/%d/transaction", baseURL, userID), bytes.NewBuffer(body))
	req.Header.Set("Source-Type", "game")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK from transaction, got %v", resp.Status)
	}

	// Step 3: Get balance again
	resp, err = http.Get(fmt.Sprintf("%s/user/%d/balance", baseURL, userID))
	if err != nil {
		t.Fatalf("Failed to fetch updated balance: %v", err)
	}
	defer resp.Body.Close()

	var after map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&after)

	finalBalance, ok := after["balance"].(string)
	if !ok {
		t.Fatalf("Final balance not found or invalid")
	}

	var afterVal float64
	fmt.Sscanf(finalBalance, "%f", &afterVal)

	// Step 4: Check if balance increased by 5.00
	expectedIncrease := 5.00
	if (afterVal - before) < expectedIncrease {
		t.Errorf("Expected balance to increase by %.2f, got %.2f → %.2f", expectedIncrease, before, afterVal)
	}
}
