package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type apiConfig struct {
	DB            *sql.DB
	InterviewRepo *interview.Repository
	UserRepo      *user.Repository
	TokenRepo     *token.Repository
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

	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}

	interviewRepo := interview.NewRepository(db)
	userRepo := user.NewRepository(db)
	tokenRepo := token.NewRepository(db)

	apiCfg := &apiConfig{
		DB:            db,
		InterviewRepo: interviewRepo,
		UserRepo:      userRepo,
		TokenRepo:     tokenRepo,
	}

	mux.HandleFunc("/api/users", apiCfg.usersHandler)
	mux.HandleFunc("/api/auth/login", apiCfg.loginHandler)
	// handlerInterviews is takes a token and interview preferences, and create new Interview resource.
	mux.HandleFunc("/api/interviews", apiCfg.interviewsHandler)
	mux.HandleFunc("/api/auth/token", apiCfg.refreshTokensHandler)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(http.ListenAndServe(":8080", enableCors(mux)))

}
