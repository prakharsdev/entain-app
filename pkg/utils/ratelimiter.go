package utils

import (
	"net"
	"net/http"
	"os"
	"strconv"
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

	defaultRPS   = getEnvAsInt("RATE_LIMIT_RPS", 30)
	defaultBurst = getEnvAsInt("RATE_LIMIT_BURST", 60)
)

func init() {
	go cleanupExpiredClients()
}

// getEnvAsInt reads an environment variable and returns it as an integer.
// If not set or invalid, it falls back to the provided default value.
func getEnvAsInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
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
		limiter := rate.NewLimiter(rate.Limit(defaultRPS), defaultBurst)
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
