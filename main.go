package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/interview"
)

type apiConfig struct {
	DB            *sql.DB
	InterviewRepo *interview.Repository
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

	interviewRepo := interview.NewRepository(db)

	apiCfg := &apiConfig{
		DB:            db,
		InterviewRepo: interviewRepo,
	}

	mux.HandleFunc("/api/users", apiCfg.usersHandler)
	mux.HandleFunc("/api/login", apiCfg.loginHandler)
	//handlerInterviews is going to take a map of user data, interview preferences, and return the first question from ChatGPT
	mux.HandleFunc("/api/interviews", apiCfg.interviewsHandler)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(http.ListenAndServe(":8080", enableCors(mux)))

}
