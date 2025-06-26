package utils

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ClientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu          sync.Mutex
	clients     = make(map[string]*ClientLimiter)
	cleanupFreq = time.Minute * 5
)

func init() {
	go cleanupExpiredClients()
}

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	client, exists := clients[ip]
	if !exists {
		limiter := rate.NewLimiter(2, 3) // 30 RPS burst of 60
		clients[ip] = &ClientLimiter{limiter, time.Now()}
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}
func cleanupExpiredClients() {
	for {
		time.Sleep(cleanupFreq)
		mu.Lock()
		for ip, client := range clients {
			if time.Since(client.lastSeen) > time.Minute*10 {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
