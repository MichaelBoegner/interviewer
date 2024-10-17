package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	DB            *sql.DB
	InterviewRepo interview.InterviewRepo
	UserRepo      user.UserRepo
	TokenRepo     token.TokenRepo
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
	log.SetFlags(log.LstdFlags | log.Llongfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

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

	mux.Handle("/api/users/{id}", middleware.GetContext(http.HandlerFunc(apiCfg.usersHandler)))
	mux.Handle("/api/auth/login", middleware.GetContext(http.HandlerFunc(apiCfg.loginHandler)))
	mux.Handle("/api/interviews", middleware.GetContext(http.HandlerFunc(apiCfg.interviewsHandler)))
	mux.Handle("/api/auth/token", middleware.GetContext(http.HandlerFunc(apiCfg.refreshTokensHandler)))

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(http.ListenAndServe(":8080", enableCors(mux)))

}
