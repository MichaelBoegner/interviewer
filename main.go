package main

import (
	"log"
	"net/http"
	"os"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"

	"github.com/joho/godotenv"
)

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// In production, this should be a specific domain. For development, we check the environment
		if os.Getenv("ENV") == "production" {
			log.Println("Using production")
			w.Header().Set("Access-Control-Allow-Origin", "https://interviewer-ui.vercel.app")
		} else {
			log.Println("NOT using production")
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env.dev"); err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
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
	conversationRepo := conversation.NewRepository(db)
	openAI := &chatgpt.OpenAIClient{}

	handler := handlers.NewHandler(interviewRepo, userRepo, tokenRepo, conversationRepo, openAI)

	mux.Handle("/api/users/", http.HandlerFunc(handler.UsersHandler))
	mux.Handle("/api/auth/login", http.HandlerFunc(handler.LoginHandler))

	mux.Handle("/api/interviews", middleware.GetContext(http.HandlerFunc(handler.InterviewsHandler)))
	mux.Handle("/api/conversations/", middleware.GetContext(http.HandlerFunc(handler.ConversationsHandler)))
	mux.Handle("/api/auth/token", middleware.GetContext(http.HandlerFunc(handler.RefreshTokensHandler)))

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(http.ListenAndServe(":8080", enableCors(mux)))

}
