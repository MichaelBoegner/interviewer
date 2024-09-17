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

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins, or set specific domain
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		next.ServeHTTP(w, r)
	})
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	// srv := &http.Server{
	// 	Addr:    ":" + port,
	// 	Handler: mux,
	// }

	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := &apiConfig{
		DB: db,
	}

	mux.HandleFunc("/api/users", apiCfg.handlerUsers)
	mux.HandleFunc("/api/login", apiCfg.handlerLogin)
	//handlerInterviews is going to take a map of user data, interview preferences, and return the first question from ChatGPT
	mux.HandleFunc("/api/interviews", apiCfg.handlerInterviews)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(http.ListenAndServe(":8080", enableCors(mux)))

}
