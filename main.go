package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/michaelboegner/interviewer/database"
)

type apiConfig struct {
	DB *sql.DB
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := &apiConfig{
		DB: db,
	}

	mux.HandleFunc("/api/users", apiCfg.handlerUsers)
	mux.HandleFunc("/api/login", apiCfg.handlerLogin)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

}
