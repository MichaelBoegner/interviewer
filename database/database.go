package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func StartDB() (*Database, error) {
	connStr := "user=michaelboegner dbname=interviewerio sslmode=disable"

	// Open the connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ensure the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the database successfully!")

	database := &Database{
		DB: db,
	}

	return database, nil

}
