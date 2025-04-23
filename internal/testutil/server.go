package testutil

import (
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/mocks"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/mailer"
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
	mailer := mailer.Mailer{}
	billing := billing.Billing{}

	handler := handlers.NewHandler(interviewRepo, userRepo, tokenRepo, conversationRepo, billing, mailer, openAI, db)

	TestMux = http.NewServeMux()
	TestMux.Handle("/api/users", http.HandlerFunc(handler.CreateUsersHandler))
	TestMux.Handle("/api/auth/login", http.HandlerFunc(handler.LoginHandler))
	TestMux.Handle("/api/auth/request-reset", http.HandlerFunc(handler.RequestResetHandler))
	TestMux.Handle("/api/auth/reset-password", http.HandlerFunc(handler.ResetPasswordHandler))

	TestMux.Handle("/api/users/", middleware.GetContext(http.HandlerFunc(handler.GetUsersHandler)))
	TestMux.Handle("/api/interviews", middleware.GetContext(http.HandlerFunc(handler.InterviewsHandler)))
	TestMux.Handle("/api/conversations/create/", middleware.GetContext(http.HandlerFunc(handler.CreateConversationsHandler)))
	TestMux.Handle("/api/conversations/append/", middleware.GetContext(http.HandlerFunc(handler.AppendConversationsHandler)))
	TestMux.Handle("/api/auth/token", middleware.GetContext(http.HandlerFunc(handler.RefreshTokensHandler)))
	TestMux.Handle("/api/payment/checkout", middleware.GetContext(http.HandlerFunc(handler.CreateCheckoutSessionHandler)))
	TestMux.Handle("/health", http.HandlerFunc(handler.HealthCheckHandler))

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
