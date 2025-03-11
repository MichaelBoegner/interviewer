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

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	var connStr string
	connStr = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbName, sslMode)

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
