package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"entain-app/internal/db"
	"entain-app/internal/user"
)

func TestHandleBalance_WithGorillaMux(t *testing.T) {
	// Step 1: Connect to DB (real one via docker)
	db.InitDB()
	db.RunMigrations()

	// Step 2: Setup Gorilla Mux with path param
	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/balance", user.HandleBalance)

	ts := httptest.NewServer(router)
	defer ts.Close()

	// Step 3: Make GET request
	resp, err := http.Get(ts.URL + "/user/1/balance")
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Step 4: Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Step 5: Assert values
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if result["userId"] != float64(1) {
		t.Errorf("Expected userId 1, got %v", result["userId"])
	}
}
