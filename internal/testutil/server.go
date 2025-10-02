package testutil

import (
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/michaelboegner/interviewer/billing"
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

func InitTestServer(logger *slog.Logger) (*handlers.Handler, error) {
	logger.Info("Initializing test database connection...")

	db, err := database.StartDB()
	if err != nil {
		logger.Error("Failed to connect to test database", "error", err)
	}

	logger.Info("Database connected successfully.")

	interviewRepo := interview.NewRepository(db)
	userRepo := user.NewRepository(db)
	tokenRepo := token.NewRepository(db)
	conversationRepo := conversation.NewRepository(db)
	billingRepo := billing.NewRepository(db)
	openAI := mocks.NewMockOpenAIClient()
	mailer := mocks.NewMockMailer()
	billing, err := billing.NewBilling(logger)
	if err != nil {
		logger.Error("billing.NewBilling failed", "error", err)
		return nil, err
	}

	handler := handlers.NewHandler(interviewRepo, userRepo, tokenRepo, conversationRepo, billingRepo, billing, mailer, openAI, db)

	TestMux = http.NewServeMux()
	TestMux.Handle("/api/users", http.HandlerFunc(handler.CreateUsersHandler))
	TestMux.Handle("/api/auth/login", http.HandlerFunc(handler.LoginHandler))
	TestMux.Handle("/api/auth/github", http.HandlerFunc(handler.GithubLoginHandler))
	TestMux.Handle("/api/auth/request-verification", http.HandlerFunc(handler.RequestVerificationHandler))
	TestMux.Handle("/api/auth/check-email", http.HandlerFunc(handler.CheckEmailHandler))
	TestMux.Handle("/api/auth/request-reset", http.HandlerFunc(handler.RequestResetHandler))
	TestMux.Handle("/api/auth/reset-password", http.HandlerFunc(handler.ResetPasswordHandler))
	TestMux.Handle("/api/auth/token", http.HandlerFunc(handler.RefreshTokensHandler))
	TestMux.Handle("/api/webhooks/billing", http.HandlerFunc(handler.BillingWebhookHandler))
	TestMux.Handle("/api/jd", http.HandlerFunc(handler.JDInputHandler))
	TestMux.Handle("/health", http.HandlerFunc(handler.HealthCheckHandler))

	TestMux.Handle("/api/users/",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.GetUsersHandler),
			),
		),
	)
	TestMux.Handle("/api/users/delete/",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.DeleteUserHandler),
			),
		),
	)
	TestMux.Handle("/api/interviews",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.InterviewsHandler),
			),
		),
	)
	TestMux.Handle("/api/interviews/",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.Method {
					case http.MethodGet:
						handler.GetInterviewHandler(w, r)
					case http.MethodPatch:
						handler.UpdateInterviewStatusHandler(w, r)
					default:
						handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
					}
				}),
			),
		),
	)
	TestMux.Handle("/api/conversations/create/",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.CreateConversationsHandler),
			),
		),
	)
	TestMux.Handle("/api/conversations/append/",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.AppendConversationsHandler),
			),
		),
	)
	TestMux.Handle("/api/conversations/",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.GetConversationHandler),
			),
		),
	)
	TestMux.Handle("/api/payment/checkout",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.CreateCheckoutSessionHandler),
			),
		),
	)
	TestMux.Handle("/api/payment/cancel",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.CancelSubscriptionHandler),
			),
		),
	)
	TestMux.Handle("/api/payment/resume",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.ResumeSubscriptionHandler),
			),
		),
	)
	TestMux.Handle("/api/payment/change-plan",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.ChangePlanHandler),
			),
		),
	)
	TestMux.Handle("/api/user/dashboard",
		middleware.GetContext(
			middleware.ValidateUserActive(userRepo)(
				http.HandlerFunc(handler.DashboardHandler),
			),
		),
	)

	logger.Info("Starting in-memory test server...")

	TestServer = httptest.NewServer(TestMux)
	TestServerURL = TestServer.URL

	return handler, nil
}

func StopTestServer() {
	if TestServer != nil {
		TestServer.Close()
	}
}
