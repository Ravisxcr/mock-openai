package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Global DB instance
var DB *sql.DB

// InitDB configures the SQLite connection and prepares the schema.
func InitDB(dataSourceName string) {
	var err error

	// 1. Open Connection
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to database: %v", err)
	}

	// 2. Connection Pool Tuning
	// SQLite only supports one writer at a time. Tuning these prevents "database is locked" errors.
	DB.SetMaxOpenConns(1) // SQLite handles concurrency better with 1 open connection
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(time.Hour)

	// 3. Verify Connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Critical: Database is unreachable: %v", err)
	}

	// 4. Initialize Schema
	createSchema()

	log.Printf("Database initialized at: %s", dataSourceName)
}

func createSchema() {
	query := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		ip TEXT,
		status INTEGER,
		latency TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(query); err != nil {
		log.Fatalf("Critical: Failed to create logs table: %v", err)
	}
}

// SaveLog writes a single request entry to SQLite.
func SaveLog(method, path, ip string, status int, latency string) {
	// Use a prepared statement internally via Exec for better performance/security
	query := `INSERT INTO request_logs (method, path, ip, status, latency) VALUES (?, ?, ?, ?, ?)`

	_, err := DB.Exec(query, method, path, ip, status, latency)
	if err != nil {
		log.Printf("DB Error: Failed to save request log: %v", err)
	}
}

// CloseDB gracefully shuts down the database connection.
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}
