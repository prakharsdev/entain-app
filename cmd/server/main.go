package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"entain-app/internal/db"
	"entain-app/internal/user"
	"entain-app/pkg/utils"
)

func main() {
	// Step 0: Initialize structured logger
	utils.InitLogger()
	utils.Logger.Info("Logger initialized")

	// Step 1: Connect to the database
	db.InitDB()
	utils.Logger.Info("Connected to DB")

	// Step 2: Run migrations + seed
	db.RunMigrations()
	utils.Logger.Info("Migrations completed")

	// Step 3: Setup HTTP router
	r := mux.NewRouter()

	// Main API routes
	r.HandleFunc("/user/{userId}/transaction", user.HandleTransaction).Methods("POST")
	r.HandleFunc("/user/{userId}/balance", user.HandleBalance).Methods("GET")

	// Health check route
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.DB.Ping(); err != nil {
			http.Error(w, `{"status":"unhealthy","database":"disconnected"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","database":"connected"}`))
	}).Methods("GET")

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Step 4: Apply middleware stack (panic recovery → logging → rate limiting)
	stacked := utils.ChainMiddlewares(r,
		utils.RecoverMiddleware,
		utils.LoggingMiddleware,
		utils.RateLimitMiddleware,
	)

	// Step 5: Configure HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: stacked,
	}

	// Step 6: Run server in a goroutine
	go func() {
		utils.Logger.Info("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.WithError(err).Fatal("ListenAndServe error")
		}
	}()

	// Step 7: Wait for SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	utils.Logger.Info("Gracefully shutting down...")

	// Step 8: Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		utils.Logger.WithError(err).Fatal("Server forced to shutdown")
	}

	utils.Logger.Info("Server exited cleanly")
}
