package utils

import (
	"net/http"
	"time"
)

// LoggingMiddleware logs method, path, and duration using structured logs
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			Logger.WithFields(map[string]interface{}{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": time.Since(start).Milliseconds(), // in ms
			}).Info("Handled request")
		}()
		next.ServeHTTP(w, r)
	})
}

// RecoverMiddleware recovers from panics and logs error with stack trace
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				Logger.WithField("panic", rec).Error("Panic recovered")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ChainMiddlewares applies multiple middleware functions in order
func ChainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}
