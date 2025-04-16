package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func StartDB() (*sql.DB, error) {
	fmt.Println("StartDB firing")

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	var connStr string

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)
	} else {
		connStr = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbName, sslMode)
	}

	supabaseConnStr := os.Getenv("DATABASE_URL")
	if supabaseConnStr != "" {
		connStr = supabaseConnStr
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the database successfully!")

	return db, nil
}
