package testutil

import (
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/mocks"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

var (
	TestMux       *http.ServeMux
	TestServer    *httptest.Server
	TestServerURL string
)

func InitTestServer() *handlers.Handler {
	log.Println("Initializing test database connection...")

	db, err := database.StartDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	log.Println("Database connected successfully.")

	interviewRepo := interview.NewRepository(db)
	userRepo := user.NewRepository(db)
	tokenRepo := token.NewRepository(db)
	conversationRepo := conversation.NewRepository(db)
	openAI := &mocks.MockOpenAIClient{}

	handler := handlers.NewHandler(interviewRepo, userRepo, tokenRepo, conversationRepo, openAI)

	TestMux = http.NewServeMux()
	TestMux.Handle("/api/users/", http.HandlerFunc(handler.UsersHandler))
	TestMux.Handle("/api/auth/login", http.HandlerFunc(handler.LoginHandler))

	TestMux.Handle("/api/interviews", middleware.GetContext(http.HandlerFunc(handler.InterviewsHandler)))
	TestMux.Handle("/api/auth/token", middleware.GetContext(http.HandlerFunc(handler.RefreshTokensHandler)))
	TestMux.Handle("/api/conversations/", middleware.GetContext(http.HandlerFunc(handler.ConversationsHandler)))

	log.Println("Starting in-memory test server...")

	TestServer = httptest.NewServer(TestMux)
	TestServerURL = TestServer.URL

	return handler
}

func StopTestServer() {
	if TestServer != nil {
		TestServer.Close()
	}
}
