package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func StartDB() (*sql.DB, error) {
	fmt.Printf("StartDB firing\n")

	// Use environment variables for database connection
	// In local development, these will come from .env
	// In production, they will be set on the hosting platform
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	// If any of these are not set, use defaults or log errors
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "michaelboegner" // Default local development user
	}
	if dbName == "" {
		dbName = "interviewerio" // Default local development database
	}
	if sslMode == "" {
		sslMode = "disable" // Local development typically has SSL disabled
	}

	// Construct the connection string
	var connStr string
	if dbPassword != "" {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)
	} else {
		connStr = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbName, sslMode)
	}

	// For direct Supabase connection string (alternative method)
	supabaseConnStr := os.Getenv("DATABASE_URL")
	if supabaseConnStr != "" {
		connStr = supabaseConnStr
	}

	// Open the connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the database successfully!")

	return db, nil
}
