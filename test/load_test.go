package test

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestLoadTransactionEndpoint(t *testing.T) {
	const (
		url     = "http://localhost:8080/user/1/transaction"
		payload = `{"state":"win", "amount":"1.00", "transactionId":"txn_perf_%d"}`
		rps     = 25 // simulate 25 requests per second
	)

	client := &http.Client{}
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < rps; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			body := bytes.NewBufferString(fmt.Sprintf(payload, i))

			req, err := http.NewRequest("POST", url, body)
			if err != nil {
				t.Errorf("Request error: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Source-Type", "game")

			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("HTTP error: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Unexpected status: %d", resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Processed %d requests in %v", rps, time.Since(start))
}
