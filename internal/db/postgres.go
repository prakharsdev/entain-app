package db

import (
	"database/sql"
	"time"

	"entain-app/configs"
	"entain-app/pkg/utils"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	cfg := configs.LoadDBConfig()

	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		utils.Logger.WithError(err).Fatal("Failed to open DB connection")
	}

	// Set DB connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := conn.Ping(); err != nil {
		utils.Logger.WithError(err).Fatal("Could not connect to the database")
	}

	utils.Logger.Info("Connected to PostgreSQL database")
	DB = conn
}
