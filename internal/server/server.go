package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/database"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/mailer"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer() (*Server, error) {
	mux := http.NewServeMux()

	db, err := database.StartDB()
	if err != nil {
		return nil, fmt.Errorf("failed to start DB: %w", err)
	}

	interviewRepo := interview.NewRepository(db)
	userRepo := user.NewRepository(db)
	tokenRepo := token.NewRepository(db)
	conversationRepo := conversation.NewRepository(db)
	billingRepo := billing.NewRepository(db)
	openAI := chatgpt.NewOpenAI()
	mailer := mailer.NewMailer()
	billing, err := billing.NewBilling()
	if err != nil {
		log.Printf("billing.NewBilling failed: %v", err)
		return nil, err
	}

	handler := handlers.NewHandler(interviewRepo, userRepo, tokenRepo, conversationRepo, billingRepo, billing, mailer, openAI, db)

	mux.Handle("/api/users", http.HandlerFunc(handler.CreateUsersHandler))
	mux.Handle("/api/auth/login", http.HandlerFunc(handler.LoginHandler))
	mux.Handle("/api/auth/request-verification", http.HandlerFunc(handler.RequestVerificationHandler))
	mux.Handle("/api/auth/check-email", http.HandlerFunc(handler.CheckEmailHandler))
	mux.Handle("/api/auth/request-reset", http.HandlerFunc(handler.RequestResetHandler))
	mux.Handle("/api/auth/reset-password", http.HandlerFunc(handler.ResetPasswordHandler))
	mux.Handle("/api/webhooks/billing", http.HandlerFunc(handler.BillingWebhookHandler))

	mux.Handle("/api/users/", middleware.GetContext(http.HandlerFunc(handler.GetUsersHandler)))
	mux.Handle("/api/interviews", middleware.GetContext(http.HandlerFunc(handler.InterviewsHandler)))
	mux.Handle("/api/interviews/", middleware.GetContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetInterviewHandler(w, r)
		case http.MethodPatch:
			handler.UpdateInterviewStatusHandler(w, r)
		default:
			handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))
	mux.Handle("/api/conversations/create/", middleware.GetContext(http.HandlerFunc(handler.CreateConversationsHandler)))
	mux.Handle("/api/conversations/append/", middleware.GetContext(http.HandlerFunc(handler.AppendConversationsHandler)))
	mux.Handle("/api/conversations/", middleware.GetContext(http.HandlerFunc(handler.GetConversationHandler)))
	mux.Handle("/api/auth/token", middleware.GetContext(http.HandlerFunc(handler.RefreshTokensHandler)))
	mux.Handle("/api/payment/checkout", middleware.GetContext(http.HandlerFunc(handler.CreateCheckoutSessionHandler)))
	mux.Handle("/api/payment/cancel", middleware.GetContext(http.HandlerFunc(handler.CancelSubscriptionHandler)))
	mux.Handle("/api/payment/resume", middleware.GetContext(http.HandlerFunc(handler.ResumeSubscriptionHandler)))
	mux.Handle("/api/payment/change-plan", middleware.GetContext(http.HandlerFunc(handler.ChangePlanHandler)))
	mux.Handle("/api/user/dashboard", middleware.GetContext(http.HandlerFunc(handler.DashboardHandler)))
	mux.Handle("/health", http.HandlerFunc(handler.HealthCheckHandler))

	return &Server{mux: mux}, nil
}

func (s *Server) StartServer() {
	log.Printf("Serving files from %s on port: %s\n", ".", "8080")
	log.Fatal(http.ListenAndServe(":8080", middleware.EnableCors(s.mux)))
}
